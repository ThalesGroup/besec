import React, { Fragment, useState } from 'react'
import {
    InputAdornment,
    TextField,
    Typography,
    Divider,
    Grid,
    makeStyles,
    Theme,
    createStyles,
    Tooltip,
    Button,
} from '@material-ui/core'
import { Error, LooksOne, LooksTwo, Looks3, Looks4, Looks5, NoteAdd, Description } from '@material-ui/icons'
import Markdown from '../common/MaterialMarkdown'

import * as api from '../../client'
import { useDispatch } from 'react-redux'
import { respondToTask, selectIsPlanReadonly, selectTaskResponses, selectPlanRevision } from '../../redux/plans'
import clsx from 'clsx'
import { useSelector } from '../../redux'

import { TwoColumns } from './TwoColumns'
import { QuestionList } from './Questions'
import { IssuesList } from './Issues'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        root: { padding: theme.spacing(1) },
        maturityIcon: { position: 'relative', top: '6px' },
        taskHeader: { display: 'flex', justifyContent: 'space-between', padding: theme.spacing(0, 1, 0) },
        '@keyframes bounce': {
            '0%': {
                transform: 'scale3d(1.0, 1.0, 1.0)',
            },
            '50%': {
                transform: 'scale3d(1.4, 1.4, 1.0)',
            },
            '100%': {
                transform: 'scale3d(1.0, 1.0, 1.0)',
            },
        },
        star: {
            fontSize: 30,
            color: '#ccd6dd',
            cursor: 'pointer',
            userSelect: 'none',
            '&:hover': { color: '#ffac33' },
            '&.active': { color: '#ffac33', animation: '$bounce 0.4s ease-in-out' },
            '&.disabled': { cursor: 'default', pointerEvents: 'none' },
        },
        references: {
            display: 'flex',
            marginTop: theme.spacing(2),
        },
        pullNext: { marginBottom: theme.spacing(-1) }, // the padding around lists is excessive in some situations
    })
)

// Tasks renders the provided array of tasks; if a planId is specified, each task has its plan data and is editable
export function Tasks(props: {
    tasks: { task: api.Task; practiceId: string }[]
    planId?: string
    revisionId?: string
}) {
    return (
        <Grid>
            {props.tasks.map((tm) => (
                <Fragment key={tm.task.id}>
                    <Task
                        task={tm.task}
                        practiceId={tm.practiceId}
                        planId={props.planId}
                        revisionId={props.revisionId}
                    />
                    <Divider /> {/* This would be better as CSS styling of the list items */}
                </Fragment>
            ))}
        </Grid>
    )
}

function Task(props: { task: api.Task; practiceId: string; planId?: string; revisionId?: string }) {
    const classes = useStyles()
    const readOnly = useSelector((state) =>
        props.planId ? selectIsPlanReadonly(state, props.planId, props.revisionId) : true
    )

    const title = (
        <div className={classes.taskHeader}>
            <Typography variant={props.planId ? 'subtitle1' : 'h6'} gutterBottom={props.planId ? false : true}>
                {props.task.title}
                <MaturityIcon level={props.task.level} />
            </Typography>
            {props.planId && (
                <Priority
                    planId={props.planId}
                    revisionId={props.revisionId}
                    practiceId={props.practiceId}
                    taskId={props.task.id}
                />
            )}
        </div>
    )

    const questions = (
        <QuestionList
            qs={props.task.questions}
            practiceId={props.practiceId}
            planId={props.planId}
            revisionId={props.revisionId}
        />
    )

    if (props.planId) {
        const responses = (
            <>
                <IssuesList
                    practiceId={props.practiceId}
                    planId={props.planId}
                    revisionId={props.revisionId}
                    taskId={props.task.id}
                    readOnly={readOnly}
                    links
                />
                {questions}
            </>
        )

        return (
            <div id={props.task.id} className={classes.root}>
                <TwoColumns
                    title={title}
                    left={responses}
                    right={
                        <>
                            <Markdown source={props.task.description} paraVariant="body2" />
                            <References
                                practiceId={props.practiceId}
                                planId={props.planId}
                                revisionId={props.revisionId}
                                taskId={props.task.id}
                                readOnly={readOnly}
                            />
                        </>
                    }
                />
            </div>
        )
    } else {
        return (
            <div id={props.task.id} className={classes.root}>
                {title}
                <Markdown source={props.task.description} paraVariant="body1" />
                <Typography variant="subtitle2" className={classes.pullNext}>
                    Questions:
                </Typography>
                {questions}
            </div>
        )
    }
}

function References(props: {
    planId: string
    revisionId?: string
    practiceId: string
    taskId: string
    readOnly: boolean
}) {
    const dispatch = useDispatch()
    const references = useSelector((state) =>
        selectTaskResponses(state, props.planId, props.practiceId, props.taskId, props.revisionId)
    )?.references
    const [adding, setAdding] = useState(false)
    const classes = useStyles()

    const RefToolTip = (props: { children: JSX.Element }) => (
        <Tooltip
            title="References point out where an activity is implemented, for example: config files; process documents; code; tool URLs; etc."
            enterDelay={200}
            placement="top-start"
        >
            {props.children}
        </Tooltip>
    )

    let content
    if (props.readOnly) {
        if (references) {
            content = (
                <>
                    <RefToolTip>
                        <Description />
                    </RefToolTip>
                    <div>
                        {/* Markdown needs to be wrapped in its own div, as it renders as a collection of <p> tags */}
                        <Markdown source={references} paraVariant="body2" />
                    </div>
                </>
            )
        } else {
            return null
        }
    } else {
        if (references || adding) {
            content = (
                <TextField
                    defaultValue={references}
                    multiline
                    fullWidth
                    variant="outlined"
                    label="References"
                    autoFocus={adding}
                    onBlur={(event) => {
                        dispatch(
                            respondToTask({
                                id: props.planId,
                                pId: props.practiceId,
                                tId: props.taskId,
                                response: { references: event.target.value },
                            })
                        )
                        setAdding(false)
                    }}
                    InputProps={{
                        startAdornment: (
                            <InputAdornment position="start">
                                <RefToolTip>
                                    <Description />
                                </RefToolTip>
                            </InputAdornment>
                        ),
                    }}
                />
            )
        } else {
            content = (
                <RefToolTip>
                    <Button
                        size="small"
                        startIcon={<NoteAdd /* also PostAdd */ />}
                        onClick={() => {
                            setAdding(true)
                        }}
                    >
                        References
                    </Button>
                </RefToolTip>
            )
        }
    }
    return <div className={classes.references}>{content}</div>
}

function Priority(props: { planId: string; revisionId?: string; practiceId: string; taskId: string }) {
    const classes = useStyles()
    const dispatch = useDispatch()
    const isPriority = useSelector(
        (state) =>
            selectPlanRevision(state, props.planId, props.revisionId).responsesById?.[props.practiceId]?.[props.taskId]
                .priority ?? false
    )
    const readOnly = useSelector((state) => selectIsPlanReadonly(state, props.planId, props.revisionId))

    return (
        <Tooltip
            title={isPriority ? 'This task has been marked as a priority to work on' : 'Mark this task as a priority'}
            enterDelay={200}
            placement="top-start"
        >
            <span
                className={clsx([classes.star, isPriority && 'active', readOnly && 'disabled'])}
                onClick={() => {
                    // won't be fired when readOnly
                    dispatch(
                        respondToTask({
                            id: props.planId,
                            pId: props.practiceId,
                            tId: props.taskId,
                            response: { priority: !isPriority },
                        })
                    )
                }}
            >
                ★
            </span>
        </Tooltip>
    )
}

function MaturityIcon(props: { level: number }) {
    const classes = useStyles()
    let Icon
    switch (props.level) {
        case 1:
            Icon = LooksOne
            break
        case 2:
            Icon = LooksTwo
            break
        case 3:
            Icon = Looks3
            break
        case 4:
            Icon = Looks4
            break
        case 5:
            Icon = Looks5
            break
        default:
            Icon = Error
    }
    return (
        <Tooltip title={`Maturity level ${props.level}`} enterDelay={150} placement="top-start">
            <Icon className={classes.maturityIcon} color="primary" />
        </Tooltip>
    )
}
