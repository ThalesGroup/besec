import React, { useMemo } from 'react'
import { Paper, Typography, makeStyles, Theme, createStyles } from '@material-ui/core'
import { blue } from '@material-ui/core/colors'
import Markdown from '../common/MaterialMarkdown'

import * as api from '../../client'
import clsx from 'clsx'

import { Tasks } from './Tasks'
import { TwoColumns } from './TwoColumns'
import { QuestionList } from './Questions'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        root: { padding: theme.spacing(1) },
        practiceNotesQuestions: {
            margin: theme.spacing(1, -1, 3, -1),
            backgroundColor: blue[50],
        },
        notApplicable: {
            zIndex: 1,
            opacity: 0.5,
        },
    })
)

export interface PracticeProps {
    practice: api.IPractice
    planId?: string // Practice serves two different purposes depending on if planId is provided or not. With it, it renders a practice in the Plan questionnaire. Without it, it renders a read-only view of the practice.
    revisionId?: string // can only be specified if planId is also specified. If provided, a readonly view of the revision is provided.
    applicable?: boolean // defaults to true
}
export default function Practice(props: PracticeProps) {
    const classes = useStyles()

    let applicable
    if (props.applicable === undefined) {
        applicable = true // applicable only makes sense in the Plan context; otherwise we don't want to hide anything
    } else {
        applicable = props.applicable
    }

    const practiceQs = useMemo(
        () => <PracticeNotesQuestions practice={props.practice} planId={props.planId} revisionId={props.revisionId} />,
        [props.practice, props.planId, props.revisionId]
    )
    // The tasks don't change when applicability changes, just the Paper they live on - memoizing them to avoid re-rendering, which is slow for large practices
    const tasks = useMemo(
        () => (
            <Tasks
                tasks={props.practice.tasks.map((t) => ({ task: t, practiceId: props.practice.id }))}
                planId={props.planId}
                revisionId={props.revisionId}
            />
        ),
        [props.practice, props.planId, props.revisionId]
    )

    return (
        <div id={props.practice.id}>
            <Typography variant="h5" gutterBottom>
                {props.practice.name}
            </Typography>
            {practiceQs}
            <Paper className={clsx(!applicable && classes.notApplicable)}>{tasks}</Paper>
        </div>
    )
}

function PracticeNotesQuestions(props: { practice: api.IPractice; planId?: string; revisionId?: string }) {
    const classes = useStyles()

    const title = (
        <Typography variant="subtitle1">
            <em>Practice notes and applicability</em>
        </Typography>
    )

    let notes
    if (props.practice.notes && props.practice.notes !== '') {
        notes = <Markdown source={props.practice.notes} paraVariant={props.planId ? 'body2' : 'body1'} />
    }

    let content
    if (props.practice.questions) {
        const qList = (
            <QuestionList
                qs={props.practice.questions}
                practiceId={props.practice.id}
                planId={props.planId}
                revisionId={props.revisionId}
            />
        )

        if (props.planId) {
            // Considered doing collapsed notes a bit like this: https://github.com/cht8687/react-text-collapse/blob/master/src/ReactTextCollapse.js
            // But they fit ok in two columns like the task descriptions.
            content = <TwoColumns title={title} left={qList} right={notes ? notes : <></>} />
        } else {
            content = (
                <>
                    {title}
                    <div className={classes.root}>{notes}</div>
                    <Typography variant="subtitle1">Practice applicability questions:</Typography>
                    {qList}
                </>
            )
        }
    } else if (notes) {
        content = (
            <>
                {title}
                <div className={classes.root}>{notes}</div>
            </>
        )
    }

    if (content) {
        return <Paper className={classes.practiceNotesQuestions}>{content}</Paper>
    } else {
        return null
    }
}
