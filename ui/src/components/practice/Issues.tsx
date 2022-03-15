import React, { useEffect, useState } from 'react'
import { TextField, makeStyles, Theme, createStyles, Button, IconButton, Link } from '@material-ui/core'
import { AddBox, Delete, Edit } from '@material-ui/icons'

import { useDispatch } from 'react-redux'
import { respondToTask, selectTaskResponses } from '../../redux/plans'
import { useSelector } from '../../redux'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        issue: {
            padding: theme.spacing(0, 1),
            display: 'inline-flex',
            alignItems: 'center',
        },
        issuesContainer: {
            padding: theme.spacing(0, 1),
        },
    })
)

// IssuesList renders all of the issues for a task, as short links where possible
export function IssuesList(props: {
    planId: string
    revisionId?: string
    practiceId: string
    taskId: string
    readOnly: boolean
    links: boolean
    prefix?: string
}) {
    const [newIssue, setNewIssue] = useState(false)
    const [editingIssues, setEditingIssues] = useState<string[]>([]) // we use the issue value rather than its position, as the store guarantees issue uniqueness
    const dispatch = useDispatch()
    const issues = useSelector(
        (state) => selectTaskResponses(state, props.planId, props.practiceId, props.taskId, props.revisionId)?.issues
    )
    const classes = useStyles()

    const save = (value: string, orig?: string) => {
        let newIssues
        if (issues) {
            newIssues = issues.slice(0)
            if (orig) {
                newIssues[newIssues.indexOf(orig)] = value
            } else {
                newIssues.push(value)
            }
        } else {
            newIssues = [value]
        }
        dispatch(
            respondToTask({
                id: props.planId,
                pId: props.practiceId,
                tId: props.taskId,
                response: { issues: newIssues },
            })
        )
    }

    if (props.readOnly && (!issues || issues.length === 0)) {
        return null
    }
    return (
        <div className={classes.issuesContainer}>
            {props.prefix}
            {issues?.map((issue) => {
                if (!props.readOnly && editingIssues.includes(issue)) {
                    return (
                        <EditIssue
                            key={issue}
                            initialValue={issue}
                            save={(value) => {
                                setEditingIssues(editingIssues.filter((e) => e !== issue))
                                save(value, issue)
                            }}
                        />
                    )
                } else {
                    let issueElem
                    try {
                        const url = new URL(issue)
                        const parts = url.pathname.split('/')
                        let display = url.pathname // fallback to full path
                        const last = parts.slice(-1)[0]
                        if (last !== '') {
                            display = last
                        } else if (parts.length > 2) {
                            display = parts.slice(-2)[0]
                        }
                        if (props.links) {
                            issueElem = (
                                <Link href={url.href} target="_blank" rel="noopener noreferrer">
                                    {display}
                                </Link>
                            )
                        } else {
                            issueElem = <span>{display}</span>
                        }
                    } catch (e) {
                        issueElem = <span>{issue}</span>
                    }
                    return (
                        <span key={issue} className={classes.issue}>
                            {issueElem}
                            {!props.readOnly && (
                                <IconButton
                                    size="small"
                                    onClick={() => {
                                        setEditingIssues([...editingIssues, issue])
                                    }}
                                >
                                    <Edit />
                                </IconButton>
                            )}
                            {!props.readOnly && (
                                <IconButton
                                    size="small"
                                    onClick={() => {
                                        dispatch(
                                            respondToTask({
                                                id: props.planId,
                                                pId: props.practiceId,
                                                tId: props.taskId,
                                                response: { issues: issues!.filter((e) => e !== issue) },
                                            })
                                        )
                                    }}
                                >
                                    <Delete />
                                </IconButton>
                            )}
                        </span>
                    )
                }
            })}
            {newIssue && (
                <EditIssue
                    initialValue=""
                    save={(value: string) => {
                        setNewIssue(false)
                        save(value)
                    }}
                />
            )}
            {!props.readOnly && (
                <Button size="small" startIcon={<AddBox />} onClick={() => setNewIssue(true)}>
                    Issue
                </Button>
            )}
        </div>
    )
}

function EditIssue(props: { initialValue: string; save: (value: string) => void }) {
    const [issueValue, setIssue] = useState('')
    useEffect(() => {
        setIssue(props.initialValue)
    }, [props.initialValue])

    return (
        <TextField
            value={issueValue}
            onChange={(event) => setIssue(event.target.value)}
            onBlur={() => {
                props.save(issueValue)
            }}
            onKeyPress={(ev) => {
                if (ev.key === 'Enter') {
                    props.save(issueValue)
                }
            }}
            autoFocus={true}
        />
    )
}
