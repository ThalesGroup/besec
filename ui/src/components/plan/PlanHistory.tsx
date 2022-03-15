import React, { useState, useEffect } from 'react'
import {
    Button,
    Checkbox,
    Link,
    makeStyles,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    CircularProgress,
    Popover,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
} from '@material-ui/core'
import { useDispatch } from 'react-redux'
import moment from 'moment'
import clsx from 'clsx'
import { Link as RouterLink } from 'react-router-dom'
import { diff as deepDiff, Diff } from 'deep-diff'

import { useSelector } from '../../redux'
import {
    selectPlanVersions,
    selectPlanRevision,
    PlanVersion,
    fetchPlan,
    PlanRevision,
    updateWorkingCopy,
    PlanVersionID,
    savePlan,
} from '../../redux/plans'
import { UserNameIcon } from '../Auth'
import { ErrorMsg } from '../common/common'
import { fetchPractices, selectPracticesVersionState } from '../../redux/practices'

const useStyles = makeStyles((theme) => ({
    current: { fontWeight: theme.typography.fontWeightBold as number },
}))

export function PlanHistory(props: { planId: string; currentRevision?: string }) {
    const classes = useStyles()
    const versions = useSelector((state) => selectPlanVersions(state, props.planId))
    const [showCompare, setShowCompare] = useState(false)
    const [first, setFirst] = useState<string | undefined>(
        props.currentRevision ?? (versions?.length > 0 ? versions[versions.length - 1].revisionId : undefined)
    )
    const [second, setSecond] = useState<string | undefined>()
    const [restoreVersion, setRestoreVersion] = useState<PlanVersionID | undefined>()

    if (!versions || versions.length === 0) {
        return null
    }

    const latest = versions[versions.length - 1].revisionId
    let current = props.currentRevision
    if (!current) {
        current = latest
    }

    const findVersion = (id: string) => versions.find((v) => v.revisionId === id)!

    const toggle = (id: string) => {
        if (first === id) {
            setFirst(undefined)
        } else if (second === id) {
            setSecond(undefined)
        } else if (!first) {
            setFirst(id)
        } else if (!second) {
            setSecond(id)
        }
    }

    const sortedVersions = () => {
        const e = [
            { id: first!, version: findVersion(first!) },
            { id: second!, version: findVersion(second!) },
        ].sort((a, b) => a.version.time.localeCompare(b.version.time))
        return {
            id1: e[0].id,
            v1: e[0].version,
            id2: e[1].id,
            v2: e[1].version,
        }
    }

    return (
        <div>
            <Button
                disabled={!first || !second}
                onClick={() => {
                    setShowCompare(true)
                }}
            >
                Compare selected revisions
            </Button>
            <Table aria-label="revision history">
                <TableHead>
                    <TableRow>
                        <TableCell>Compare</TableCell>
                        <TableCell>Published</TableCell>
                        <TableCell>Changed by</TableCell>
                        <TableCell>Actions</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {versions
                        .map((v) => (
                            <TableRow key={v.revisionId}>
                                <TableCell>
                                    <Checkbox
                                        checked={[first, second].includes(v.revisionId)}
                                        onChange={() => {
                                            toggle(v.revisionId)
                                        }}
                                    ></Checkbox>
                                </TableCell>
                                <TableCell>
                                    <RouterLink
                                        to={
                                            '/plan/' +
                                            props.planId +
                                            (v.revisionId === latest ? '' : '?revision=' + v.revisionId)
                                        }
                                        className={clsx(v.revisionId === current && classes.current)}
                                    >
                                        {formatTime(v.time)}
                                    </RouterLink>
                                </TableCell>
                                <TableCell>
                                    <UserNameIcon
                                        user={v.author}
                                        className={clsx(v.revisionId === current && classes.current)}
                                    />
                                </TableCell>
                                <TableCell>
                                    {v.revisionId !== latest && (
                                        <Link
                                            href="#"
                                            onClick={() => {
                                                setRestoreVersion(v)
                                            }}
                                        >
                                            Restore
                                        </Link>
                                    )}
                                </TableCell>
                            </TableRow>
                        ))
                        .reverse()}
                </TableBody>
            </Table>

            {showCompare && (
                <Compare
                    planId={props.planId}
                    {...sortedVersions()}
                    close={() => {
                        setShowCompare(false)
                    }}
                />
            )}
            <Restore planId={props.planId} version={restoreVersion} onClose={() => setRestoreVersion(undefined)} />
        </div>
    )
}

function Compare(props: {
    planId: string
    id1: string
    v1: PlanVersion
    id2: string
    v2: PlanVersion
    close: () => void
}) {
    const [r1, r2] = useSelector((state) => [
        selectPlanRevision(state, props.planId, props.id1),
        selectPlanRevision(state, props.planId, props.id2),
    ])
    const [p1, p2] = useSelector((state) =>
        [r1?.practicesVersion, r2?.practicesVersion].map((v) => selectPracticesVersionState(state, v))
    )
    const dispatch = useDispatch()

    // fetch things we need: an existing plan; a template; practice versions
    useEffect(() => {
        // The fetch actions won't do unnecessary work, so always dispatch them
        dispatch(fetchPlan(props.planId, { revisionId: props.id1, fetchResponses: true }))
        dispatch(fetchPlan(props.planId, { revisionId: props.id2, fetchResponses: true }))
        r1?.practicesVersion && dispatch(fetchPractices({ version: r1.practicesVersion }))
        r2?.practicesVersion && dispatch(fetchPractices({ version: r2.practicesVersion }))
    }, [props, r1, r2, dispatch])

    const friendlyName = (pId: string, qtId: string) => {
        if (qtId === 'undefined') {
            // Practice doesn't exist any more (don't ask why it's a string)
            return ''
        }
        // The element may only be present in one of the practice sets
        for (const p of [p1, p2]) {
            const practice = p?.byId?.[pId]
            if (!practice) {
                continue
            }

            const task = practice.tasks.find((t) => t.id === qtId)
            if (task) {
                return task.title
            } else if (practice.questions?.find((q) => q.id === qtId)) {
                return "Practice question '" + qtId + "'"
            } else {
                const t = practice.tasks.find((t) => t.questions.find((q) => q.id === qtId))
                if (t) {
                    if (t.questions.length === 1) {
                        return t.title
                    } else {
                        return t.title + ": '" + qtId + "'"
                    }
                }
            }
        }
        return qtId
    }

    let contents
    if (
        r1?.detailsFetched &&
        r2?.detailsFetched &&
        r1?.responsesReady &&
        r2?.responsesReady &&
        p1?.fetched &&
        p2?.fetched
    ) {
        const d = revdiff(r1, r2)
        const differences: JSX.Element[] = []
        if (r1.practicesVersion !== r2.practicesVersion) {
            differences.push(
                <TableRow key="version">
                    <TableCell>Practices Version</TableCell>
                    <TableCell></TableCell>
                    <TableCell>
                        <RouterLink to={`/practices/?version=${r1.practicesVersion}`}>{r1.practicesVersion}</RouterLink>
                    </TableCell>
                    <TableCell>
                        <RouterLink to={`/practices/?version=${r2.practicesVersion}`}>{r2.practicesVersion}</RouterLink>
                    </TableCell>
                </TableRow>
            )
        }
        if (d.details) {
            d.details.forEach((diff) => {
                switch (diff.kind) {
                    case 'N':
                    case 'E': {
                        const elem = diff.path?.[1].charAt(0).toUpperCase() + diff.path?.[1].slice(1)
                        differences.push(
                            <TableRow key={diff.path?.toString()}>
                                <TableCell>Details</TableCell>
                                <TableCell>{elem}</TableCell>
                                <TableCell>{diff.kind === 'E' && diff.lhs.toString()}</TableCell>
                                <TableCell>{diff.rhs.toString()}</TableCell>
                            </TableRow>
                        )
                        break
                    }
                    case 'A':
                        differences.push(
                            <TableRow key={diff.path?.toString()}>
                                <TableCell>Details</TableCell>
                                <TableCell>Projects</TableCell>
                                <TableCell>{diff.item.kind === 'D' && diff.item.lhs.toString()}</TableCell>
                                <TableCell>{diff.item.kind === 'N' && '+ ' + diff.item.rhs.toString()}</TableCell>
                            </TableRow>
                        )
                        break
                }
            })
        }
        if (d.responses) {
            Object.entries(d.responses).forEach(([pId, taskDiffs]) => {
                Object.entries(taskDiffs).forEach(([qtId, taskDiff]) => {
                    differences.push(
                        <TableRow key={pId + qtId}>
                            <TableCell>{p1.byId?.[pId]?.name ?? p2.byId?.[pId]?.name ?? pId}</TableCell>
                            <TableCell>{friendlyName(pId, qtId)}</TableCell>
                            {taskCompare(taskDiff)}
                        </TableRow>
                    )
                })
            })
        }
        contents = (
            <Table aria-label="differences">
                <TableHead>
                    <TableRow>
                        <TableCell>Section</TableCell>
                        <TableCell>Element</TableCell>
                        <TableCell>{formatTime(props.v1.time)}</TableCell>
                        <TableCell>{formatTime(props.v2.time)}</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>{differences}</TableBody>
            </Table>
        )
    } else if (r1?.fetchErrMsg) {
        contents = <ErrorMsg message={'Error fetching revision: ' + r1.fetchErrMsg} />
    } else if (r2?.fetchErrMsg) {
        contents = <ErrorMsg message={'Error fetching revision: ' + r2.fetchErrMsg} />
    } else if (p1?.fetchErrMsg) {
        contents = <ErrorMsg message={'Error fetching practices version: ' + p1.fetchErrMsg} />
    } else if (p2?.fetchErrMsg) {
        contents = <ErrorMsg message={'Error fetching practices version: ' + p2.fetchErrMsg} />
    } else {
        contents = <CircularProgress />
    }
    return (
        <Popover
            PaperProps={{ style: { minWidth: '70%', maxWidth: '90%', maxHeight: '90%', overflowY: 'scroll' } }}
            open
            onClose={props.close}
            anchorReference="anchorPosition"
            anchorPosition={{ top: 150, left: 250 }}
        >
            <div style={{ minHeight: 400 }}>{contents}</div>
        </Popover>
    )
}

function taskCompare(diff: TaskDiff) {
    const lhs: JSX.Element[] = []
    const rhs: JSX.Element[] = []

    const processMissing = (d: Diff<PlanRevision, PlanRevision>, answers: boolean) => {
        let missingType
        if (d.path?.length === 2) {
            missingType = 'Practice'
        } else if (d.path?.length === 3) {
            missingType = 'Task'
        }
        if (!missingType) {
            return false
        }
        let present, content, absent
        if (d.kind === 'N') {
            present = rhs
            content = d.rhs
            absent = lhs
        } else if (d.kind === 'D') {
            present = lhs
            content = d.lhs
            absent = rhs
        } else {
            console.warn(`Got unexpected diff type '${d.kind}' at ${d.path}`)
            return
        }
        present.push(<p key={d.path!.join(',')}>{JSON.stringify(content)}</p>)
        if (!answers) {
            // only record this the first time, i.e. when processing tasks
            absent.push(
                <p
                    key={d.path!.join(',')}
                >{`<${missingType} ID doesn't exist in this revision's version of the practices>`}</p>
            )
        }
        return true
    }

    diff.task.forEach((d) => {
        if (processMissing(d, false)) {
            return
        } else if (d.path?.length !== 4) {
            console.warn('Got unexpected diff path: ' + d.path)
        } else {
            const elem = d.path[3].charAt(0).toUpperCase() + d.path[3].slice(1)
            if (d.kind === 'D' || d.kind === 'E') {
                lhs.push(<p key={elem}>{elem + ': ' + (d.lhs?.toString() ?? '')}</p>)
            }
            if (d.kind === 'E' || d.kind === 'N') {
                rhs.push(<p key={elem}>{elem + ': ' + (d.rhs?.toString() ?? '')}</p>)
            }
            if (d.kind === 'A') {
                if (elem !== 'Issues') {
                    console.warn('Unexpected array diff at ' + d.path)
                    return
                }
                if (d.item.kind === 'D' || d.item.kind === 'E') {
                    lhs.push(<p key={elem}>{elem + ': ' + d.item.lhs.toString()}</p>)
                }
                if (d.item.kind === 'E' || d.item.kind === 'N') {
                    rhs.push(<p key={elem}>{elem + ': ' + d.item.rhs.toString()}</p>)
                }
            }
        }
    })

    diff.answers.forEach((d) => {
        if (processMissing(d, true)) {
            return
        } else if (d.kind !== 'E') {
            console.warn(`Got unexpected diff type '${d.kind}' at ${d.path}`)
            return
        } else if (d.path?.length !== 4) {
            console.warn('Got unexpected diff path: ' + d.path)
            return
        } else {
            let prefix = ''
            if (d.path[3] === 'notes') {
                prefix = 'Notes: '
            }
            lhs.push(<p key={prefix}>{prefix + (d.lhs?.toString() ?? '')}</p>)
            rhs.push(<p key={prefix}>{prefix + (d.rhs?.toString() ?? '')}</p>)
        }
    })

    return [<TableCell key={'r1'}>{lhs}</TableCell>, <TableCell key={'r2'}>{rhs}</TableCell>]
}

type RevDeepDiff = Diff<PlanRevision, PlanRevision>
type TaskDiff = {
    task: RevDeepDiff[]
    answers: RevDeepDiff[]
}
function revdiff(r1: PlanRevision, r2: PlanRevision) {
    const detailsDiff: RevDeepDiff[] = []
    const responsesDiff: {
        [practice: string]: {
            [qtid: string]: TaskDiff // id might be for a question or a task
        }
    } = {}
    deepDiff(r1, r2)?.forEach((d) => {
        if (d.path?.[0] === 'details') {
            // don't report differences in the calculated maturity
            if (d.path?.[1] !== 'maturity') {
                detailsDiff.push(d)
            }
        } else if (d.path?.[0]) {
            const pId = d.path![1] as string
            if (pId) {
                const qtId = d.path![2] as string
                if (!(pId in responsesDiff)) {
                    responsesDiff[pId] = {}
                }
                if (!(qtId in responsesDiff[pId])) {
                    responsesDiff[pId][qtId] = { task: [], answers: [] }
                }

                if (d.path[0] === 'responsesById') {
                    responsesDiff[pId][qtId].task.push(d)
                } else if (d.path[0] === 'answersById') {
                    responsesDiff[pId][qtId].answers.push(d)
                } else {
                    console.warn('Got a difference at unexpected path under a practice: ' + d.path)
                }
            } else {
                if (d.path[0] !== 'practicesVersion') {
                    console.warn('Got a difference at unexpected path ' + d.path)
                }
            }
        } // else - metadata, discard. possibly an error condition.
    })

    return { details: detailsDiff, responses: responsesDiff }
}

function formatTime(time?: string) {
    if (!time) return ''
    return moment(time).local().format('YYYY-MM-DD HH:mm')
}

function Restore(props: { planId: string; version?: PlanVersionID; onClose: () => void }) {
    const dispatch = useDispatch()
    const restore = () => {
        dispatch(updateWorkingCopy({ planId: props.planId, revId: props.version!.revisionId }))
        dispatch(savePlan(props.planId))
        props.onClose()
    }

    return (
        <Dialog open={!!props.version} onClose={props.onClose} aria-describedby="restore-dialog-description">
            <DialogContent>
                <DialogContentText id="restore-dialog-description">
                    Restore the revision created at {formatTime(props.version?.time)} by {props.version?.author?.name}?
                </DialogContentText>
                <DialogContentText>
                    (You can always restore the current revision if you change your mind.)
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.onClose} color="primary">
                    Cancel
                </Button>
                <Button onClick={restore} color="primary" autoFocus>
                    OK
                </Button>
            </DialogActions>
        </Dialog>
    )
}
