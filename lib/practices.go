package lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/go-openapi/strfmt"
	"github.com/imdario/mergo"
	"github.com/santhosh-tekuri/jsonschema/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

// Practice is the internal representation of a BeSec practice, including all of its tasks and questions
type Practice struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Questions []Question `json:"questions"`
	Tasks     []Task     `json:"tasks"`
	Level0    Level0     `json:"level0"`
	Page      string     `json:"page"`      // Practice page URL
	Condition string     `json:"condition"` // how to interpret a practice's qualifying questions
	Notes     string     `json:"notes"`
}

// Task represents an individual task within a Practice
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Level       uint8      `json:"level"`
	Questions   []Question `json:"questions"`
}

// Question represents both maturity questions and qualifying questions
type Question struct {
	ID    string `json:"id"` // mandatory for qualifying questions
	Text  string `json:"text"`
	NA    bool   `json:"na"`
	Other bool   `json:"other"`
}

// Level0 describes the state of a project that has not met Level 1 for the practice
type Level0 struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

// practiceDef represents a practice definition file. It is similar to, and ultimately converted into, a Practice
type practiceDef struct {
	Remove          bool // used for deltas to indicate that a base practice shouldn't be used
	ID              string
	Name            string
	Questions       []Question
	Tasks           []string
	TaskDefinitions map[string]Task `yaml:"taskDefinitions"` // though in the file, tasks dont have an ID field, so all of the tasks will have blank IDs
	Level0          Level0
	Page            string
	Condition       string
	Notes           string
}

// FromDef converts the file representation of a practice into the internal/API representation
func (p *Practice) FromDef(md practiceDef) error {
	p.ID = md.ID
	p.Name = md.Name
	p.Questions = md.Questions
	p.Level0 = md.Level0
	p.Page = md.Page
	p.Condition = md.Condition
	p.Notes = md.Notes

	p.Tasks = []Task{}
	for _, tID := range md.Tasks {
		t, ok := md.TaskDefinitions[tID]
		t.ID = tID
		if !ok {
			return fmt.Errorf("no definition found for task %v from the tasks list", tID)
		}
		p.Tasks = append(p.Tasks, t)
	}
	return nil
}

// Validate is a dummy function, required for the generated API code.
// Internally, practices are always valid, so we don't need to validate them prior to sending them.
func (p *Practice) Validate(formats interface{}) error {
	return nil
}

// ContextValidate is a dummy function, required for the generated API code.
// Internally, practices are always valid, so we don't need to validate them prior to sending them.
func (p *Practice) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// EvaluateCondition evaluates the practice's condition with the provided named values
func (p Practice) EvaluateCondition(parameters map[string]interface{}) (bool, error) {
	e, err := govaluate.NewEvaluableExpression(p.Condition)
	if err != nil {
		return false, err
	}
	result, err := e.Evaluate(parameters)
	if err != nil {
		return false, err
	}
	switch result := result.(type) {
	case bool:
		return result, nil
	default:
		return false, fmt.Errorf("Practice %v condition evaluation didn't result in a boolean! %v", p.ID, result)
	}
}

// CheckConstraints checks additional constraints on:
//  - practice conditions
//  - task and question IDs
//  - task maturity levels
// and also populates implicit question IDs
//nolint:gocognit
func (p *Practice) CheckConstraints() error {
	if len(p.Questions) > 0 { // there must be a valid condition if there are any practice questions
		if _, err := govaluate.NewEvaluableExpression(p.Condition); err != nil {
			return fmt.Errorf("Failed to parse condition '%v'", p.Condition)
		}
	}

	qIds := make(map[string]bool)
	checkQIds := func(qs []Question) error {
		for _, q := range qs {
			// Check Tasks with multiple questions all have question IDs
			if q.ID == "" && len(qs) > 1 {
				return fmt.Errorf("question is missing ID, but isn't a solo question in a task: %v", q.Text)
			}
			// Check Question IDs are unique
			if q.ID != "" && qIds[q.ID] {
				return fmt.Errorf("question ID %v is not unique within the practice", q.ID)
			}
			qIds[q.ID] = true
		}
		return nil
	}

	tIds := make(map[string]bool)
	levels := make(map[uint8]bool)
	for _, t := range p.Tasks {
		// Check Task IDs are unique
		if tIds[t.ID] {
			return fmt.Errorf("task ID %v is repeated", t.ID)
		}

		// Populate implicit question IDs
		if (len(t.Questions) == 1) && (t.Questions[0].ID == "") {
			t.Questions[0].ID = t.ID
		}

		if res := checkQIds(t.Questions); res != nil {
			return res
		}

		levels[t.Level] = true
	}

	if !levels[4] {
		return fmt.Errorf("practice %v doesn't have a level 4 task - the order levels must be populated is 4,1,2,3", p.ID)
	}
	if levels[2] && !levels[1] {
		return fmt.Errorf("practice %v has a level 2 task but no level 1 task - the order levels must be populated is 4,1,2,3", p.ID)
	}
	if levels[3] && !levels[2] {
		return fmt.Errorf("practice %v has a level 3 task but no level 2 task - the order levels must be populated is 4,1,2,3", p.ID)
	}

	return checkQIds(p.Questions)
}

// TasksByLevel returns all of the tasks in the practice indexed by their maturity level
func (p Practice) TasksByLevel() map[uint8][]Task {
	levels := make(map[uint8][]Task)
	for _, t := range p.Tasks {
		tasks, ok := levels[t.Level]
		if !ok {
			tasks = []Task{}
		}
		levels[t.Level] = append(tasks, t)
	}
	return levels
}

// TaskFromID returns the task with the corresponding id and true, or false if it wasn't found
func (p Practice) TaskFromID(id string) (task Task, found bool) {
	for _, t := range p.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

type practicePaths struct {
	practice string
	delta    string
	basename string
}

// ParsePracticesDir parses all of the yaml files in the parser's practices & delta dir
func (pp *PracticeParser) ParsePracticesDir() ([]Practice, error) {
	var paths []practicePaths
	var gatherYaml filepath.WalkFunc
	matchedDeltas := map[string]bool{} // delta files that have the same name as a base practice

	camelCase := regexp.MustCompile(`^[a-z]+[a-zA-Z0-9]+.yaml$`)
	gatherYaml = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matchedDeltas[path] {
			// We've already processed this file - it's a delta to a base practice
			return nil
		}
		if filepath.Ext(path) == ".yaml" {
			if !camelCase.MatchString(filepath.Base(path)) {
				return fmt.Errorf("practice file %v must be camelCase - start with a-z and only contain alphanumeric characters", path)
			}
			id := strings.TrimSuffix(filepath.Base(path), ".yaml")
			p := practicePaths{practice: path, basename: id}
			inBase, _ := filepath.Match(filepath.Join(pp.dir, "/*"), path)
			if inBase && pp.deltaDir != "" {
				// we are in the base dir, check if there's a matching delta file
				deltaPath := filepath.Join(pp.deltaDir, info.Name())
				if _, err := pp.fs.Stat(deltaPath); err == nil {
					p.delta = deltaPath
					matchedDeltas[deltaPath] = true
					log.Infof("Found delta for practice %v", id)
				}
			}
			paths = append(paths, p)
		}
		return nil
	}

	err := afero.Walk(pp.fs, pp.dir, gatherYaml)
	if err != nil {
		log.Fatalf("Failed to walk the practices directory: %v", err)
	}
	if pp.deltaDir != "" {
		err = afero.Walk(pp.fs, pp.deltaDir, gatherYaml)
		if err != nil {
			log.Fatalf("Failed to walk the delta directory: %v", err)
		}
	}
	return pp.parsePractices(paths)
}

// ParsePractices parses the listed practices
func (pp *PracticeParser) parsePractices(practicePaths []practicePaths) ([]Practice, error) {
	practices := []Practice{}
	for _, paths := range practicePaths {
		practice, err := pp.parsePractice(paths.practice, paths.delta, paths.basename)
		if err != nil {
			return []Practice{}, err
		}
		if practice != nil {
			practices = append(practices, *practice)
		}
	}
	return practices, nil
}

// readDef parses a yaml file into a practiceDef
func (pp *PracticeParser) readDef(path string, id string, delta bool) (practiceDef, error) {
	var practiceYaml []byte
	file, err := pp.fs.Open(path)
	if err != nil {
		return practiceDef{}, err
	}
	if practiceYaml, err = ioutil.ReadAll(file); err != nil {
		return practiceDef{}, err
	}

	if !delta {
		// First validate that the parsed file complies with the spec
		// This is useful because the yaml package will assign the zero value to any missing fields
		// Don't attempt schema validation of delta files - it doesn't need to comply
		// (we could create a schema for a delta file, where everything is optional)

		// Use a custom yaml unmarshaller that ensures all map keys are strings, not arbitrary nodes.
		// Technically this is not true for all yaml, but it is for our practices and we need it like this to
		// be able to use a json schema to validate it.
		var ifPractice interface{}
		err = unmarshalYaml(practiceYaml, &ifPractice)
		if err != nil {
			return practiceDef{}, fmt.Errorf("Failed to parse %v: %v", path, err)
		}
		if err = pp.schema.ValidateInterface(ifPractice); err != nil {
			return practiceDef{}, fmt.Errorf("Failed to validate %v: %v", path, err)
		}
	}

	// Now actually load the practice
	// We have to parse it twice, as schema validation doesn't work on the Practice struct
	var def practiceDef
	err = yaml.UnmarshalStrict(practiceYaml, &def)
	if err != nil {
		if delta {
			return practiceDef{}, fmt.Errorf("Error unmarshalling delta practice %v: %v", path, err)
		}
		return practiceDef{}, fmt.Errorf("Error unmarshalling practice %v, but validation against the schema succeeded! %v", path, err)
	}
	if def.ID != "" && def.ID != id {
		// ignore if the ID is missing - this must be a delta file
		return practiceDef{}, fmt.Errorf("Practice at path %v must have ID %v, but got %v", path, id, def.ID)
	}

	return def, nil
}

// ParsePractice parses the yaml files at practicePath and deltaPath into a Practice
// The return value may be nil, if the delta practice removes the base practice
// It verifies that the practice ID matches the provided ID
func (pp *PracticeParser) parsePractice(practicePath string, deltaPath string, id string) (*Practice, error) {
	def, err := pp.readDef(practicePath, id, false)
	if err != nil {
		return &Practice{}, err
	}

	if deltaPath != "" {
		var deltaDef practiceDef
		deltaDef, err = pp.readDef(deltaPath, id, true)
		if err != nil {
			return &Practice{}, err
		}

		err = mergo.Merge(&def, deltaDef, mergo.WithOverride)
		if err != nil {
			return &Practice{}, fmt.Errorf("failed to merge delta practice at %v onto base practice: %v", deltaPath, err)
		}
	}

	if def.Remove {
		return nil, nil
	}

	practice := Practice{}
	err = practice.FromDef(def)
	if err != nil {
		return &Practice{}, fmt.Errorf("Error parsing %v: %v", id, err)
	}

	// There are some properties that JSON Schema can't validate, and implicit values need to be made explicit
	if err = practice.CheckConstraints(); err != nil {
		return &Practice{}, fmt.Errorf("Practice %v passed schema-validation but doesn't meet additional constraints: %v", practicePath, err)
	}

	return &practice, nil
}

// PracticeParser parses practices within a specific directory, applying deltas from another dir if specified
type PracticeParser struct {
	dir      string
	deltaDir string
	schema   *jsonschema.Schema
	fs       afero.Fs
}

// NewPracticeParser creates a PracticeParser for the specified directory
// if fs is nil, the OS fs is used
func NewPracticeParser(practicesDir string, deltaDir, schemaPath string, fs afero.Fs) PracticeParser {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	// Set up the schema compiler
	schemaFile, err := fs.Open(schemaPath)
	if err != nil {
		log.Fatalf("Error loading practice schema - if you've specified a non-default practices-dir, try passing --schema-file=practices/schema.json. Error was: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	if err = compiler.AddResource(schemaPath, schemaFile); err != nil {
		panic(err)
	}
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		log.Panicf("Error parsing practice schema! Check --schema-file. %v", err)
	}

	return PracticeParser{schema: schema, dir: practicesDir, deltaDir: deltaDir, fs: fs}
}
