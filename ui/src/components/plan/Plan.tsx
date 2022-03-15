import React, { useEffect, useMemo, useState } from 'react'
import { useDispatch, shallowEqual } from 'react-redux'
import { CircularProgress, makeStyles, Theme, createStyles, Grid, Paper, Typography } from '@material-ui/core'
import Moment from 'moment'
import { Prompt } from 'react-router'
import { RouteComponentProps, Link } from 'react-router-dom'

import Toc from '../common/Toc'
import { useSelector } from '../../redux'
import {
    fetchPractices,
    selectLatestPracticesVersion,
    selectPracticeById,
    selectPracticesVersionState,
} from '../../redux/practices'
import {
    createNewPlan,
    NEWPLAN,
    fetchPlan,
    removeNewPlan,
    makeSelectUnansweredPracticeQuestions as makeSelectUnansweredPracticeQPractices,
    makeSelectPlanDetailsValid,
    makeSelectApplicablePracticeIDs,
    makeSelectUnansweredPractices,
    updateWorkingCopy,
    selectPlanStateById,
    selectPlanDeleted,
    selectPlanExists,
} from '../../redux/plans'
import PlanActions from './PlanActions'
import Practice from '../practice/Practice'
import PlanDetails from './PlanDetails'
import { Refresh } from '../common/Refresh'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        progress: { position: 'fixed', left: '50%', top: '25%' },
        paper: { padding: theme.spacing(2) },
    })
)

function getProjectIds(params: URLSearchParams): string[] | null {
    const projects = params.get('projects')
    if (!projects || projects.length === 0) {
        return null
    }
    return projects.split(',')
}
// Plan renders a new plan if projectId is specified, or the latest revision of an existing plan if planID is provided
// For new plans, on save the user will get redirected to the now-saved project
export default function Plan({ match, location, history }: RouteComponentProps<{ planId: string; section: string }>) {
    const classes = useStyles()

    const id = match.params.planId
    const section = match.params.section
    const params = new URLSearchParams(location.search)
    const projectIds = getProjectIds(params)
    const revParam = params.get('revision')
    const revisionId = revParam ? revParam : undefined

    const dispatch = useDispatch()

    // Fetch all the plan properties we need, rather than the whole plan, as that leads to a re-render on every answer entry
    const {
        planExists,
        revisionExists,
        planDeleted,
        responsesReady,
        practicesVersion,
        fetchErrMsg,
        revisionFetchErrMsg,
        saved,
        detailsFetched,
        savedPlanId,
        isDeleting,
        isFetching,
        isSaving,
        committed,
    } = useSelector((state) => {
        const planExists = selectPlanExists(state, id)
        const plan = selectPlanStateById(state, id)
        const revision = revisionId ? plan?.revisionsById[revisionId] : plan?.workingCopy
        const revisionExists = revisionId ? !!revision : planExists
        return {
            planExists,
            revisionExists,
            planDeleted: selectPlanDeleted(state, id),
            fetchErrMsg: plan?.fetchErrMsg,
            revisionFetchErrMsg: revision && revision.fetchErrMsg,
            saved: planExists && plan.saved,
            savedPlanId: plan?.id ?? '',
            isDeleting: planExists && plan.isDeleting,
            isSaving: planExists && plan.isSaving,
            isFetching: revisionExists && (revision.fetchingDetails || revision.fetchingResponses),
            responsesReady: revisionExists && revision.responsesReady,
            practicesVersion: id === NEWPLAN ? selectLatestPracticesVersion(state) : revision?.practicesVersion,
            detailsFetched: revisionExists && revision.detailsFetched,
            committed: revisionExists && revision.details.committed,
        }
    }, shallowEqual)

    const practiceState = useSelector((state) => selectPracticesVersionState(state, practicesVersion))
    const practice = useSelector((state) => selectPracticeById(state, practicesVersion, section))

    const templateId = params.get('template')
    const template = useSelector((state) => {
        if (!templateId || id !== NEWPLAN) return null // ignore any template parameter when we aren't creating a new plan
        const exists = selectPlanExists(state, templateId)
        return {
            id: templateId,
            exists,
            deleted: selectPlanDeleted(state, id),
            plan: exists ? selectPlanStateById(state, templateId) : undefined,
        }
    })

    // We need to specify shallowEqual as a comparison function here - the returned arrays will be not be === to each other
    // after any answer is updated, even though most answer updates will not lead to a change in the array contents.
    const selectUnansweredPracticeQPractices = useMemo(makeSelectUnansweredPracticeQPractices, [])
    const unansweredPracticeQPractices = useSelector(
        (state) => selectUnansweredPracticeQPractices(state, id, revisionId),
        shallowEqual
    )
    const selectApplicablePractices = useMemo(makeSelectApplicablePracticeIDs, [])
    const applicablePractices = useSelector((state) => selectApplicablePractices(state, id, revisionId), shallowEqual)
    const selectUnansweredPractices = useMemo(makeSelectUnansweredPractices, [])
    const unansweredPractices = useSelector((state) => selectUnansweredPractices(state, id, revisionId), shallowEqual)
    const selectPlanDetailsValid = useMemo(makeSelectPlanDetailsValid, [])
    const planDetailsValid = useSelector((state) => selectPlanDetailsValid(state, id, revisionId))

    // Remove the new plan, if it's hanging around from earlier
    useEffect(() => {
        if (id === NEWPLAN) {
            dispatch(removeNewPlan())
        }
    }, [id, dispatch]) // only run this on first render

    // For existing plans, blat any unsaved changes to the working copy
    // This will also trigger unnecessarily when we load a plan for the first time, but that doesn't matter
    const [refreshedWorking, setRefreshedWorking] = useState(false)
    useEffect(() => {
        if (planExists && !saved && !refreshedWorking) {
            dispatch(updateWorkingCopy({ planId: id }))
            setRefreshedWorking(true)
        }
    }, [id, planExists, saved, dispatch, refreshedWorking, setRefreshedWorking])

    // initialize a new plan
    useEffect(() => {
        if (id === NEWPLAN && !planExists) {
            if (practiceState?.fetched) {
                // the practices have been fetched, create an empty plan from them
                const today = Moment().format('YYYY-MM-DD')

                if (template) {
                    if (template.plan?.workingCopy.responsesReady) {
                        dispatch(createNewPlan(today, projectIds, template.id))
                    }
                } else {
                    dispatch(createNewPlan(today, projectIds))
                }
            }
        }
    })

    // fetch things we need: an existing plan; a template; the right practice definitions
    useEffect(() => {
        // The fetch actions won't do unnecessary work, so always dispatch them
        if (id) {
            dispatch(fetchPlan(id, { revisionId, fetchResponses: true, fetchHistory: true }))
        }

        // Need to fetch the template if one has been specified (and hence we're creating a new plan)
        if (template) {
            dispatch(fetchPlan(template.id, { fetchResponses: true }))
        }

        // Only fetch the practices once we know what version of them we need
        if (practicesVersion) {
            dispatch(fetchPractices({ version: practicesVersion }))
        }
    }, [id, template, practicesVersion, revisionId, dispatch])

    // Navigate to a newly created plan on save
    useEffect(() => {
        if (id === NEWPLAN && savedPlanId && savedPlanId !== NEWPLAN) {
            // replace not push, as we never want to browse back to re-create a plan we've just created.
            history.replace('/plan/' + savedPlanId)
            dispatch(removeNewPlan())
        }
    }, [id, savedPlanId, history, dispatch])

    // Navigate away on successful plan deletion
    useEffect(() => {
        if (!planExists && planDeleted) {
            history.push('/plans')
        }
    }, [history, planDeleted, planExists])

    let content
    let error = false
    const inProgress = <CircularProgress className={classes.progress} />

    // There are a lot of different scenarios to handle.
    // If we are in-progress, usually the in-progress indicator is the only thing shown
    // For errors, we show the refresh button when there is a chance it is a transient error.
    if (!id) {
        return <p>No plan specified</p>
    } else if (fetchErrMsg) {
        error = true
        content = (
            <p>
                Error fetching plan {id}: {fetchErrMsg}
            </p>
        )
    } else if (revisionFetchErrMsg) {
        error = true
        content = (
            <p>
                Error fetching plan {id} revision {revisionId}: {revisionFetchErrMsg}
            </p>
        )
    } else if (!planExists || !practiceState) {
        if (practiceState?.fetchErrMsg) {
            error = true
            content = (
                <p>
                    Error fetching practices at version {practicesVersion}: {practiceState.fetchErrMsg}
                </p>
            )
        } else if (template?.plan?.fetchErrMsg) {
            error = true
            content = <p>Error fetching plan template: {template.plan.fetchErrMsg}</p>
        } else if (id === NEWPLAN && !projectIds) {
            return <p>When creating a new plan a project ID must be specified.</p>
        } else {
            // plan hasn't been initialized yet, the effect will trigger that
            return inProgress
        }
    } else if (revisionId && !revisionExists) {
        // revision hasn't been initialized yet, the effect will trigger that
        return inProgress
    } else if (id !== NEWPLAN && (!detailsFetched || !responsesReady)) {
        return inProgress
    } else if (!section) {
        content = (
            <PlanDetails
                id={id}
                revisionId={revisionId}
                applicablePractices={applicablePractices}
                unansweredTaskQs={unansweredPractices}
                unansweredPracticeQs={unansweredPracticeQPractices}
            />
        )
    } else if (practice) {
        // On performance: whilst the plan re-renders in response to some question answers (practice applicability change, all of practice's questions are answered),
        // re-rendering the Practice here isn't a problem because the Practice memoizes the majority of its content
        // OTOH, whenever we switch between practices it has to re-render the practice - perhaps we could cache the initial render of a practice?
        content = (
            <Practice
                practice={practice}
                planId={id}
                revisionId={revisionId}
                applicable={applicablePractices.includes(section)}
            />
        )
    } else {
        content = <p>No practice found with ID {section}.</p>
    }

    const planUrl = `/plan/${id}`
    const suffix = revisionId ? '?revision=' + revisionId : ''
    let tocLinks = [
        {
            url: planUrl + suffix,
            text: 'Plan Details',
            current: !section,
            todo: !planDetailsValid,
            applicable: true,
        },
    ]
    if (practiceState?.ids) {
        tocLinks = tocLinks.concat(
            practiceState.ids.map((practiceId) => ({
                url: planUrl + '/' + practiceId + suffix,
                text: practiceState.byId[practiceId].name,
                current: practiceId === section,
                todo:
                    unansweredPracticeQPractices.includes(practiceId) ||
                    (applicablePractices.includes(practiceId) && unansweredPractices.includes(practiceId)),
                applicable: applicablePractices.includes(practiceId),
            }))
        )
    }

    // The plan is ready to commit if none of the sections are left to do
    const planAnswered = tocLinks.reduce((prev, current) => prev && !current.todo, true)

    return (
        <>
            {/* note that this Prompt doesn't handle navigation events outside of react-router.
            They're working on it: https://github.com/ReactTraining/react-router/issues/6830#issuecomment-530574522
            My own quick attempt to use onbeforeunload within useEffect didn't work - the cleanup didn't seem to be working */}
            {!error && (
                <Prompt
                    when={!saved}
                    message={(location) =>
                        location.pathname.startsWith(`/plan/${id}`)
                            ? true
                            : 'There are unsaved changes, discard them and leave now?'
                    }
                />
            )}
            <Grid container direction="row-reverse">
                <Grid item sm={12} md={2} align-self="baseline">
                    <Toc links={tocLinks} />
                </Grid>
                <Grid item sm={12} md={10} style={{ position: 'relative' }}>
                    <Refresh
                        refreshing={isFetching}
                        refresh={() => {
                            dispatch(
                                fetchPlan(id, { revisionId, fetchResponses: true, fetchHistory: true, force: true })
                            )
                        }}
                    />
                    <Paper className={classes.paper}>
                        {revisionId && (
                            <Typography align="center" color="primary">
                                You are viewing a specified revision.{' '}
                                <Link to={'/plan/' + id + (section ? '/' + section : '')}>
                                    Return to the latest revision.
                                </Link>
                            </Typography>
                        )}
                        {content}
                        {!revisionId && !error && (
                            <PlanActions
                                id={id}
                                new={id === NEWPLAN}
                                isDeleting={isDeleting}
                                isSaving={isSaving}
                                dirty={!saved}
                                commitReady={planAnswered}
                                committed={committed}
                            />
                        )}
                    </Paper>
                </Grid>
            </Grid>
        </>
    )
}
