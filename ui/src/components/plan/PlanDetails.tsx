import React, { useState, useEffect } from 'react'
import {
    TextField,
    makeStyles,
    Theme,
    createStyles,
    Typography,
    ListItemIcon,
    ListItemText,
    List,
    Accordion,
    AccordionSummary,
    AccordionDetails,
    Tooltip,
} from '@material-ui/core'
import { useDispatch } from 'react-redux'
import moment from 'moment'
import { isEqual } from 'lodash-es'

import * as api from '../../client'
import Markdown from '../common/MaterialMarkdown'
import { ListItemLink } from '../common/common'
import { useSelector } from '../../redux'
import { setDetail, selectPlanRevision } from '../../redux/plans'
import ProjectSelect from '../project/ProjectSelect'
import { MaturityBar, CurrentMaturityBar } from '../Metrics'
import { IssuesList } from '../practice/Issues'
import { PlanHistory } from './PlanHistory'
import { ExpandMore } from '@material-ui/icons'
import { selectPracticesVersion } from '../../redux/practices'
import { selectProjectById } from '../../redux/projects'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        root: {
            '& > :first-child': { margin: theme.spacing(0, 0, 1, 0) }, // the first heading, gets a bit cramped otherwise
            '& > :not(:first-child)': { margin: theme.spacing(2, 0, 0, 0) }, // all the other direct children. This is an alternative to using the Grid's spacing feature, which doesn't work here
        },
        column: { width: 300 + theme.spacing(2), '& > *': { margin: theme.spacing(2, 0, 0, 0) } },
        colWidth: { minWidth: 300 + theme.spacing(2) },
        bold: {
            fontWeight: theme.typography.fontWeightBold as number,
        },
        subtleAccordionRoot: {
            boxShadow: 'none',
            '&:before': { display: 'none' },
        },
        subtleAccordionSummary: { padding: 0 },
    })
)

export default function PlanDetails(props: {
    id: string
    revisionId?: string
    applicablePractices: string[]
    unansweredTaskQs: string[]
    unansweredPracticeQs: string[]
}) {
    const classes = useStyles()
    const dispatch = useDispatch()
    const { exists, details, projectNames, practicesVersion } = useSelector((state) => {
        const revision = selectPlanRevision(state, props.id, props.revisionId)
        if (!revision) {
            return {
                exists: false,
                details: { date: '', committed: false, projects: [], notes: '', maturity: {} },
                projectNames: [],
                practicesVersion: '',
            }
        }
        const details = revision.details
        return {
            exists: true,
            details,
            projectNames: details.projects
                .map((id) => selectProjectById(state, id)?.attributes.name)
                .filter((v) => v) as string[],
            practicesVersion: revision.practicesVersion,
        }
    }, isEqual)

    const practices = useSelector((state) => selectPracticesVersion(state, practicesVersion))

    const priorityTasks = useSelector((state) => {
        const revision = selectPlanRevision(state, props.id, props.revisionId)
        const responses = revision?.responsesById
        if (!responses) {
            return []
        }
        return practices.reduce((tasks, p) => {
            p.practice.tasks.forEach((t) => {
                if (responses[p.id][t.id]?.priority) {
                    tasks.push({ task: t, practiceId: p.id })
                }
            })
            return tasks
        }, [] as { task: api.Task; practiceId: string }[])
    })

    const [date, setDate] = useState<string>('')
    const [notes, setNotes] = useState<string | undefined>(undefined)
    useEffect(() => {
        setDate(details.date)
        setNotes(details.notes)
    }, [details])

    if (!exists) {
        if (props.revisionId) {
            return <p>Plan revision not found</p>
        }
        return <p>Plan not found</p>
    }

    const save = (details: Partial<api.IPlanDetails>) => {
        dispatch(setDetail({ id: props.id, revisionDetails: details }))
    }

    let contents
    if (details.committed || props.revisionId) {
        contents = (
            <>
                <h2>{details.date}</h2>
                <Typography variant="body2">
                    <i>{details.committed ? 'Committed' : 'Draft'}</i>
                </Typography>
                <p>Projects: {projectNames.join(', ')}</p>
                <Markdown source={details.notes} />
            </>
        )
    } else {
        const parsedDate = moment(date, 'YYYY-M-D', true)
        contents = (
            <div className={classes.column}>
                <TextField
                    required
                    error={!parsedDate.isValid()}
                    value={date}
                    onChange={(event) => setDate(event.target.value)}
                    onBlur={() => {
                        let reformatted = date
                        if (parsedDate.isValid()) {
                            reformatted = parsedDate.format('YYYY-MM-DD')
                            setDate(reformatted)
                        }
                        save({ date: reformatted })
                    }}
                    variant="outlined"
                    label="Date"
                    className={classes.colWidth}
                    helperText="When the plan was made (YYYY-MM-DD)"
                />
                <ProjectSelect planId={props.id} projectIds={details.projects} />
                <TextField
                    value={notes}
                    onChange={(event) => setNotes(event.target.value)}
                    onBlur={() => save({ notes })}
                    variant="outlined"
                    label="Notes"
                    className={classes.colWidth}
                    multiline
                    helperText="Markdown is supported."
                />
            </div>
        )
    }

    let maturity
    const status = {
        applicablePractices: props.applicablePractices,
        unansweredTaskQs: props.unansweredTaskQs,
        unansweredPracticeQs: props.unansweredPracticeQs,
    }
    if (props.revisionId) {
        maturity = <MaturityBar maturity={details.maturity} status={status} practicesVersion={practicesVersion} />
    } else {
        maturity = <CurrentMaturityBar planId={props.id} status={status} />
    }

    const priorityTaskItems = priorityTasks.map((t) => (
        <ListItemLink
            key={t.task.id}
            to={`/plan/${props.id}/${t.practiceId}${props.revisionId ? '?revision=' + props.revisionId : ''}#${
                t.task.id
            }`}
        >
            <ListItemIcon>
                <span
                    style={{
                        fontSize: 30,
                        color: '#ffac33',
                    }}
                >
                    ★
                </span>
            </ListItemIcon>
            <ListItemText>
                {t.task.title}
                <IssuesList
                    readOnly={true}
                    links={false}
                    prefix="&nbsp;&nbsp;&nbsp;&nbsp;Issue(s): "
                    practiceId={t.practiceId}
                    planId={props.id}
                    taskId={t.task.id}
                />
            </ListItemText>
        </ListItemLink>
    ))
    let priorities
    if (priorityTaskItems.length > 0) {
        priorities = (
            <List disablePadding dense>
                <Typography variant="body1">These activities have been chosen as priorities to work on:</Typography>
                {priorityTaskItems}
            </List>
        )
    } else {
        priorities = <Typography variant="body1">No activities have been chosen as a priority to work on.</Typography>
    }

    return (
        <div className={classes.root}>
            <Typography variant="h5">Plan details</Typography>
            {contents}
            <Typography variant="h5">Maturity levels</Typography>
            {maturity}
            <Typography variant="h5">Priority activities</Typography>
            {priorities}
            <Accordion className={classes.subtleAccordionRoot}>
                <AccordionSummary
                    expandIcon={<ExpandMore />}
                    aria-controls="history-content"
                    id="history-header"
                    className={classes.subtleAccordionSummary}
                >
                    <Tooltip title="Click to show revision history" enterDelay={200} placement="top-start">
                        <Typography variant="h5">Revision history</Typography>
                    </Tooltip>
                </AccordionSummary>
                <AccordionDetails>
                    <PlanHistory planId={props.id} currentRevision={props.revisionId} />
                </AccordionDetails>
            </Accordion>
        </div>
    )
}
