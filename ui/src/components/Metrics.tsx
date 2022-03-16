import React, { useEffect, useState, useMemo } from 'react'
import { VictoryChart, VictoryBar, VictoryAxis, VictoryTheme } from 'victory'
import { useTheme, CircularProgress, Grid, Typography, Link } from '@material-ui/core'
import { useDispatch, shallowEqual } from 'react-redux'
import { RouteComponentProps, Link as RouterLink } from 'react-router-dom'

import { useSelector } from '../redux'
import {
    fetchProjects,
    selectProjectById,
    selectProjectByName,
    selectProjectIdsSorted,
    selectProjectsFetchStatus,
    SerializedProject,
} from '../redux/projects'
import { fetchPlan, makeSelectApplicablePracticeIDs, selectAnswersIfReady, selectPlanStatesById } from '../redux/plans'
import {
    ExtendedPractice,
    selectPracticesVersion,
    selectPracticesByIdIfFetched,
    fetchPractices,
} from '../redux/practices'
import * as api from '../client'
import { ErrorMsg } from './common/common'
import { Refresh } from './common/Refresh'

// The Metrics page
export default function Metrics({ match }: RouteComponentProps<{ projectName: string }>) {
    const dispatch = useDispatch()
    const { fetched, isFetching, fetchErrMsg } = useSelector(selectProjectsFetchStatus)
    const projectName = match.params.projectName
    const id = useSelector((state) => selectProjectByName(state, projectName))
    const projectIds = useSelector(selectProjectIdsSorted)
    const [refreshing, setRefreshing] = useState<boolean[]>([]) // used to trigger and record refresh of each project's data

    // Refresh projects on load
    useEffect(() => {
        dispatch(fetchProjects(true))
    }, [dispatch])

    if (!fetched) {
        if (fetchErrMsg) {
            return <ErrorMsg message={fetchErrMsg} />
        }
        return <CircularProgress />
    }

    let content
    let projects: string[]
    if (projectName) {
        if (!id) {
            projects = []
            content = <ErrorMsg message={"Couldn't find project with name " + projectName} />
        } else {
            projects = [id]
            content = (
                <ProjectMetrics
                    id={id}
                    history={true}
                    refresh={refreshing[0]}
                    refreshed={() => setRefreshing([false])}
                />
            )
        }
    } else {
        projects = projectIds

        content = (
            <>
                <Typography variant="h6" gutterBottom>
                    Practice maturity from the latest committed project plans
                </Typography>
                <Grid container spacing={4}>
                    {projects.map((id, pos) => (
                        <Grid key={id} item>
                            <ProjectMetrics
                                id={id}
                                history={false}
                                refresh={refreshing[pos]}
                                refreshed={() => {
                                    const updated = refreshing.slice(0)
                                    updated[pos] = false
                                    setRefreshing(updated)
                                }}
                            />
                        </Grid>
                    ))}
                </Grid>
            </>
        )
    }

    return (
        <div style={{ textAlign: 'center', position: 'relative' }}>
            <Refresh
                refresh={() => {
                    dispatch(fetchProjects(true))
                    setRefreshing(projects.map((_) => true))
                }}
                refreshing={isFetching || refreshing.some((e) => e)}
            />
            {content}
        </div>
    )
}

interface PlanData {
    id: string
    fetched: boolean
    fetching: boolean
    details?: api.IPlanDetails
    practicesVersion?: string
}
function ProjectMetrics(props: { id: string; history: boolean; refresh?: boolean; refreshed?: () => void }) {
    const dispatch = useDispatch()
    const project = useSelector((state) => selectProjectById(state, props.id))
    const planStates = useSelector((state) => selectPlanStatesById(state, project.plans))

    // Fetch plans (if necessary)
    useEffect(() => {
        // we only need the responses so we have the practices version used.
        // This is a design flaw - if this causes performance issues the API should be refactored to store the practices version in the revision details rather than in the responses.
        project.plans.forEach((id) => dispatch(fetchPlan(id, { fetchResponses: true })))
    }, [project.plans, dispatch])

    // Fetch plans on manual refresh
    const { refresh, refreshed } = props
    useEffect(() => {
        if (refresh) {
            project.plans.forEach((id) => dispatch(fetchPlan(id, { force: true })))
        }
    }, [refresh, project.plans, dispatch])

    const plans: PlanData[] = project.plans.map((id) => {
        const plan = planStates[id]
        if (plan) {
            return {
                id: plan.id,
                fetched: true,
                fetching: plan.workingCopy.fetchingDetails || plan.workingCopy.fetchingResponses,
                details: plan.workingCopy.details,
                practicesVersion: plan.workingCopy.practicesVersion,
            }
        } else {
            return { id, fetched: false, fetching: false }
        }
    })

    // Fetch referenced practices (if necessary)
    useEffect(() => {
        plans.forEach((plan) => {
            if (plan.practicesVersion) {
                dispatch(fetchPractices({ version: plan.practicesVersion }))
            }
        })
    }, [plans, dispatch])

    // Propagate that manual refresh of this project is complete
    // Relies upon happening after the preceding useEffect that triggers the fetches
    useEffect(() => {
        if (refresh && refreshed && plans.every((p) => !p.fetching)) {
            refreshed()
        }
    }, [plans, refresh, refreshed])

    // We need to fetch all the plans before we can know which is the latest
    // This will need server-side support at some point as it is very inefficient
    if (plans.filter((plan) => !plan.fetched).length > 0) {
        return <CircularProgress />
    }

    const committedPlans = plans
        .filter((plan) => plan.details!.committed)
        .sort((a, b) => b.details!.date.localeCompare(a.details!.date))

    if (props.history) {
        return <ProjectHistory sortedPlans={committedPlans} project={project} />
    } else {
        return <ProjectLatest sortedPlans={committedPlans} project={project} />
    }
}

function ProjectLatest(props: { sortedPlans: PlanData[]; project: SerializedProject }) {
    let content
    let latestPlan
    if (props.sortedPlans.length === 0) {
        content = <Typography variant="body2">No committed plans</Typography>
    } else {
        latestPlan = props.sortedPlans[0]

        content = (
            <MaturityBar maturity={latestPlan.details!.maturity} practicesVersion={latestPlan.practicesVersion!} />
        )
    }
    return (
        <div>
            <Link variant="h6" component={RouterLink} to={'metrics/' + props.project.attributes.name}>
                {props.project.attributes.name}
            </Link>
            {latestPlan && <PlanLink plan={latestPlan} />}
            {content}
        </div>
    )
}

function PlanLink(props: { plan: PlanData }) {
    return (
        <Link
            variant="body2"
            style={{ display: 'block' }}
            gutterBottom
            component={RouterLink}
            to={'/plan/' + props.plan.id}
        >
            {props.plan.details!.date}
        </Link>
    )
}
function ProjectHistory(props: { sortedPlans: PlanData[]; project: SerializedProject }) {
    if (props.sortedPlans.length === 0) {
        return <Typography variant="body1">No committed plans</Typography>
    }
    return (
        <Grid container spacing={2}>
            {props.sortedPlans.reverse().map((plan) => (
                <Grid key={plan.id} item>
                    <PlanLink plan={plan} />
                    <MaturityBar maturity={plan.details!.maturity} practicesVersion={plan.practicesVersion!} />
                </Grid>
            ))}
        </Grid>
    )
}

interface PracticeStatus {
    applicablePractices: string[]
    unansweredTaskQs: string[]
    unansweredPracticeQs: string[]
}
export function MaturityBar(props: {
    maturity: api.IPlanDetails['maturity']
    status?: PracticeStatus // if not provided, assume that all practices not in maturity are N/A
    practicesVersion: string
}) {
    const theme = useTheme()
    const practices = useSelector((state) => selectPracticesVersion(state, props.practicesVersion))
    const data = practices.reverse().map((p) => {
        let level = props.maturity[p.id]
        if (!level || level === -1) {
            level = 0
        }
        let label
        if (props.status) {
            if (props.status.applicablePractices.includes(p.id)) {
                if (props.status.unansweredTaskQs.includes(p.id)) {
                    label = '?'
                } else {
                    label = level
                }
            } else {
                if (props.status.unansweredPracticeQs.includes(p.id)) {
                    label = '?'
                } else {
                    label = 'N/A'
                }
            }
        } else {
            if (p.id in props.maturity) {
                label = level
            } else {
                label = 'N/A'
            }
        }
        return {
            id: p.id,
            name: p.practice.name.replace(/ /g, '\n'),
            level,
            label,
        }
    })

    if (data.length === 0) {
        return <></>
    }

    return (
        <div style={{ maxWidth: 250 }}>
            <VictoryChart
                theme={VictoryTheme.material}
                height={150 + practices.length * 30}
                width={
                    250 /* the height and width here define an aspect ratio, not an absolute size.
                           Can't get it quite right - possibly could not specify this, and instead force some category padding? */
                }
                padding={{ top: 0, right: 0, bottom: 10, left: 100 }}
                domainPadding={{ x: [0, 20] } /* the axis is weirdly offset otherwise */}
            >
                <VictoryBar
                    data={data}
                    x="name"
                    y="level"
                    labels={({ datum }) => datum.label}
                    horizontal
                    domain={[0, 5]}
                    style={{ data: { fill: theme.palette.primary.main }, labels: { color: 'white' } }}
                />
                {/* to show grid lines, if not using labels:
                <VictoryAxis
                    dependentAxis
                    tickValues={[1, 2, 3, 4]}
                    style={{
                        axis: { visibility: 'hidden' },
                        tickLabels: { visibility: 'hidden' },
                        ticks: { visibility: 'hidden' }
                    }}
                /> */}
                <VictoryAxis style={{ grid: { visibility: 'hidden' }, ticks: { visibility: 'hidden' } }} />
            </VictoryChart>
        </div>
    )
}

// CurrentMaturityBar calculates the maturity from the current revision's answers, instead of using the saved
// maturity
export function CurrentMaturityBar(props: {
    planId: string
    status?: PracticeStatus // if not provided, assume that all practices not in maturity are N/A
}) {
    const answers = useSelector((state) => selectAnswersIfReady(state, props.planId))
    const practicesById = useSelector((state) => selectPracticesByIdIfFetched(state, answers?.practicesVersion))
    const selectApplicablePractices = useMemo(makeSelectApplicablePracticeIDs, [])
    const applicablePractices = useSelector((state) => selectApplicablePractices(state, props.planId), shallowEqual)

    if (!answers) {
        return null
    }

    const maturity: api.IPlanDetails['maturity'] = {}
    if (answers.answersById) {
        applicablePractices.forEach((pId) => {
            maturity[pId] = PracticeMaturity(practicesById[pId], answers.answersById![pId])
        })
    }

    return <MaturityBar maturity={maturity} status={props.status} practicesVersion={answers.practicesVersion} />
}

// PracticeLevel returns the highest level in the given practice for which all tasks of the same or lower level are answered Yes or N/A
// Returns -1 if there are unanswered questions
// Based on the go implementation.
export function PracticeMaturity(practice: ExtendedPractice, practiceAnswers: Record<string, api.IAnswer>): number {
    const yes: boolean[] = Array(5).fill(false)
    const no: boolean[] = Array(5).fill(false)

    for (const task of practice.tasks) {
        switch (TaskResult(task, practiceAnswers)) {
            case api.Answer2.No: {
                no[task.level] = true
                break
            }
            case api.Answer2.Unanswered: {
                return -1
            }
            default: {
                //Yes & N/A
                yes[task.level] = true
                break
            }
        }
    }

    // answer is the highest 'yes' seen below the lowest no
    let lowestNo = no.indexOf(true, 1)
    if (lowestNo === -1) {
        lowestNo = 99
    }
    const highestYes = yes.slice(0, lowestNo).lastIndexOf(true)
    if (highestYes === -1) {
        return 0
    }
    return highestYes
}

// TaskResult returns No if any answer for the task is No, N/A if all are N/A, Unanswered if any answer is such,
// and Yes otherwise (i.e. at least one Yes and the rest Yes or N/A)
// exported for tests
export function TaskResult(task: api.ITask, practiceAnswers: Record<string, api.IAnswer>): api.Answer2 {
    let allNA = true
    for (const q of task.questions) {
        switch (practiceAnswers[q.id!].answer) {
            case api.Answer2.No: {
                return api.Answer2.No
            }
            case api.Answer2.Unanswered: {
                return api.Answer2.Unanswered
            }
            case api.Answer2.Yes: {
                allNA = false
                break
            }
        }
    }
    if (allNA) {
        return api.Answer2.N_A
    }
    return api.Answer2.Yes
}
