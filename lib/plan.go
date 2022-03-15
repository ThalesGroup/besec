package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
)

// AnswerVal defines possible answers to questions
type AnswerVal string

// Possible values for AnswerVal
const (
	No         AnswerVal = "No"
	Yes        AnswerVal = "Yes"
	NA         AnswerVal = "N/A"
	Unanswered AnswerVal = "Unanswered"
)

func parseAnswer(ans string) (AnswerVal, error) {
	switch strings.ToLower(ans) {
	case "yes":
		return Yes, nil
	case "no":
		return No, nil
	case "na", "n/a":
		return NA, nil
	case "unanswered":
		return Unanswered, nil
	default:
		return Unanswered, fmt.Errorf("Invalid answer '%v'", ans)
	}
}

// Validate checks the answer value is valid
func (a *Answer) Validate(formats interface{}) error {
	_, err := parseAnswer(string(a.Answer))
	return err
}

// MarshalJSON serialises a AnswerVal to JSON
func (a *AnswerVal) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*a))
}

// UnmarshalJSON creates a AnswerVal from JSON source
func (a *AnswerVal) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}
	*a, err = parseAnswer(s)
	return
}

// UnmarshalYAML creates an AnswerVal from YAML
func (a *AnswerVal) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}

	*a, err = parseAnswer(s)
	return err
}

// Plan captures the results of a questionnaire
type Plan struct {
	Details   PlanDetails   `json:"details"`
	Responses PlanResponses `json:"responses"`
}

// PlanDetails captures the high level responses
type PlanDetails struct {
	Projects  []string       `json:"projects"`
	Date      string         `json:"date"`
	Notes     string         `json:"notes"`
	Committed bool           `json:"committed"`
	Maturity  map[string]int `json:"maturity"` // keyed on practice ID, only practices with a calculable maturity are present
}

// PlanResponses captures the responses to the practices
type PlanResponses struct {
	PracticesVersion  string                      `json:"practicesVersion" yaml:"practicesVersion"`
	PracticeResponses map[string]PracticeResponse `json:"practiceResponses" yaml:"practiceResponses"` // keyed on practiceID
	practices         []Practice
}

// PracticeResponse holds the responses to the practice and task questions
type PracticeResponse struct {
	Practice map[string]Answer       `json:"practice"` // keyed on question ID
	Tasks    map[string]TaskResponse `json:"tasks"`    // keyed on task ID
}

// TaskResponse holds the answers to a task's questions and the optional extra info about a task's implementation
type TaskResponse struct {
	Answers    map[string]Answer `json:"answers"`
	Priority   bool              `json:"priority"`
	Issues     []string          `json:"issues"`
	References string            `json:"references"`
}

// Answer holds the 'hard' answer and any notes for a response to a question
type Answer struct {
	Answer AnswerVal `json:"answer"`
	Notes  string    `json:"notes"`
}

// NewPlan returns a Plan with its calculated maturity
func NewPlan(details PlanDetails, responses PlanResponses, practices []Practice) Plan {
	responses.practices = practices
	p := Plan{Details: details, Responses: responses}
	p.CalculateMaturity()
	return p
}

// CalculateMaturity sets the plan's maturity based on its answers
// Only sets a maturity level for practices that apply and are fully answered
func (plan *Plan) CalculateMaturity() {
	plan.Details.Maturity = make(map[string]int, len(plan.Responses.practices))
	for _, practice := range plan.Responses.practices {
		applies, err := plan.Responses.PracticeApplies(practice)
		if (err == nil) && applies {
			maturity, err := plan.Responses.PracticeLevel(practice)
			if err == nil {
				plan.Details.Maturity[practice.ID] = maturity
			}
		}
	}
}

// Validate that the date is in YYYY-MM-DD format and at least one project ID is supplied.
func (pd *PlanDetails) Validate(formats interface{}) error {
	if _, err := time.Parse("2006-01-02", pd.Date); err != nil {
		return fmt.Errorf("Invalid date format, must be YYYY-MM-DD: %v", pd.Date)
	}
	if len(pd.Projects) == 0 {
		return fmt.Errorf("at least one project ID must be provided")
	}
	return nil
}

// ContextValidate is required for the generated API code, but the goswagger docs don't describe its purpose.
// It is related to validating read-only properties, see https://github.com/go-swagger/go-swagger/issues/2648
func (pd *PlanDetails) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// Validate is a dummy function - we rely on the Plan being validated
func (mr *PracticeResponse) Validate(formats interface{}) error {
	return nil
}

// Validate that the responses to this Plan have answers to all questions.
// Note practice-applicability isn't considered: all questions need a response, even if it is "Unanswered"
// If the response doesn't have practices populated, no validation is done
func (responses *PlanResponses) Validate(formats interface{}) error {
	if responses.practices == nil {
		return nil
	}

	missing := responses.MissingPracticeAnswers()
	if len(missing) > 0 {
		return fmt.Errorf("missing answers for practices: %v", missing)
	}

	missing = responses.MissingAnswers(true)
	if len(missing) > 0 {
		return fmt.Errorf("missing answers for tasks: %v", missing)
	}

	return nil
}

// ContextValidate is required for the generated API code, but the goswagger docs don't describe its purpose.
// It is related to validating read-only properties, see https://github.com/go-swagger/go-swagger/issues/2648
func (responses *PlanResponses) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MissingPracticeAnswers returns a list of practice IDs which are missing responses to practice-level questions
func (responses *PlanResponses) MissingPracticeAnswers() []string {
	missing := []string{}

	for _, practice := range responses.practices {
		if len(practice.Questions) == 0 {
			continue
		}

		practiceResp, ok := responses.PracticeResponses[practice.ID]
		if !ok {
			missing = append(missing, practice.ID)
		}

		if len(practiceResp.Practice) != len(practice.Questions) {
			missing = append(missing, practice.ID)
		}
	}
	return missing
}

// PracticeApplies returns true if the plan indicates this practice applies
// If there is a missing or invalid response to a question, return an error
func (responses *PlanResponses) PracticeApplies(practice Practice) (bool, error) {
	if len(practice.Questions) == 0 {
		return true, nil
	}

	practiceResp, ok := responses.PracticeResponses[practice.ID]
	if !ok {
		return false, fmt.Errorf("missing response for practice %v", practice.ID)
	}

	for _, q := range practice.Questions {
		qResp, ok := practiceResp.Practice[q.ID]
		if !ok {
			return false, fmt.Errorf("missing response for %v.%v", practice.ID, q.ID)
		}

		if qResp.Answer == NA && !q.NA {
			return false, fmt.Errorf("%v.%v does not allow N/A as an answer", practice.ID, q.ID)
		} else if q.NA {
			log.Infof("Practice question %v.%v allows N/A as an answer, but we can't evaluate that - an N/A means the practice doesn't apply.\n", practice.ID, q.ID)
		}
	}
	params := make(map[string]interface{})
	for q, r := range practiceResp.Practice {
		switch r.Answer {
		case Yes:
			params[q] = true
		case No:
			params[q] = false
		case NA:
			return false, nil
		case Unanswered:
			return false, fmt.Errorf("Unanswered question in plan: %v", q)
		}
	}
	result, err := practice.EvaluateCondition(params)
	if err != nil {
		return false, fmt.Errorf("Failed to evaluate condition for practice %v: %v", practice.ID, err)
	}
	return result, nil
}

// TaskResult returns No if any answer for the task is No, N/A if all are N/A, Unanswered if any answer is such,
// and Yes otherwise (i.e. at least one Yes and the rest Yes or N/A)
func (responses *PlanResponses) TaskResult(practiceID string, taskID string) (AnswerVal, error) {
	taskAnswers := responses.PracticeResponses[practiceID].Tasks[taskID].Answers

	if len(taskAnswers) == 0 {
		return No, fmt.Errorf("TaskResult: no response for %v.%v", practiceID, taskID)
	}

	allNA := true
	for _, a := range taskAnswers {
		switch a.Answer {
		case No:
			return No, nil
		case Unanswered:
			return Unanswered, nil
		case Yes:
			allNA = false
		}
	}
	if allNA {
		return NA, nil
	}
	return Yes, nil
}

// PracticeLevel returns the highest level in the given practice for which all tasks of the same or lower level are answered Yes or N/A
// It only considers answers in the response - if an answer is missing, it will be as if the task doesn't exist (or didn't have that question where they have multiple qs)
// Returns an error if there are unanswered questions.
func (responses *PlanResponses) PracticeLevel(practice Practice) (int, error) {
	ts := responses.PracticeResponses[practice.ID].Tasks
	var yes, no [5]bool // no level 0 tasks, so the first entry is irrelevant

	for tID := range ts {
		t, found := practice.TaskFromID(tID)
		if !found {
			return 0, fmt.Errorf("Task in plan not found in practices: %v", tID)
		}
		res, err := responses.TaskResult(practice.ID, tID)
		if err != nil {
			return 0, err
		}
		switch res {
		case No:
			no[t.Level] = true
		case Unanswered:
			return 0, fmt.Errorf("Can't compute practice level if it has unanswered questions")
		default:
			yes[t.Level] = true
		}
	}

	// answer is the highest 'yes' seen below the lowest no
	lowestNo := 99
	for l := range no {
		if no[l] {
			lowestNo = l
			break
		}
	}
	result := 0
	for l := range yes {
		if l >= lowestNo {
			break
		}
		if yes[l] {
			result = l
		}
	}
	return result, nil
}

// MissingAnswersForPractice returns a list of task IDs in practice that don't have answers to all their questions in this plan
// ignoreUnanswered controls whether to count answers with a value of Unanswered as missing (if false) or only to
// report syntactically invalid responses that don't have any value for a task (if true).
func (responses *PlanResponses) MissingAnswersForPractice(practice Practice, ignoreUnanswered bool) (missing []string) {
	for _, t := range practice.Tasks {
		miss := true
		resp, ok := responses.PracticeResponses[practice.ID].Tasks[t.ID]
		if ok {
			for _, q := range t.Questions {
				a, ok := resp.Answers[q.ID]
				if ok {
					if ignoreUnanswered {
						miss = false
					} else {
						miss = a.Answer == Unanswered
					}
				}
			}
		}
		if miss {
			missing = append(missing, fmt.Sprintf("%v.%v", practice.ID, t.ID))
		}
	}
	return missing
}

// MissingAnswers returns a list of task IDs in practices that don't have answers to all their questions in this plan
// ignoreUnanswered controls whether to count answers with a value of Unanswered as missing (if false) or only to
// report syntactically invalid responses that don't have any value for a task (if true).
func (responses *PlanResponses) MissingAnswers(ignoreUnanswered bool) []string {
	missing := []string{}
	for _, practice := range responses.practices {
		missing = append(missing, responses.MissingAnswersForPractice(practice, ignoreUnanswered)...)
	}
	return missing
}

// ReadyToCommit return (true,[]) if all practice-level questions have been answered and applicable practices
// have answers to all of their task questions. Otherwise it returns false and a list of issues.
// Depends on the global Practices array being populated
func (responses *PlanResponses) ReadyToCommit() (bool, []string) {
	errors := []string{}
	for _, practice := range responses.practices {
		applies, err := responses.PracticeApplies(practice)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Can't tell if practice %v applies: %v", practice.ID, err))
		} else if applies {
			missing := responses.MissingAnswersForPractice(practice, false)
			if len(missing) > 0 {
				errors = append(errors, fmt.Sprintf("Missing or unanswered answers for applicable practice %v: %v", practice.ID, missing))
			}
		}
	}
	return len(errors) == 0, errors
}
