import React, { useState, useEffect, ReactElement } from 'react'
import { useDispatch } from 'react-redux'
import {
    Button,
    List,
    ListItemText,
    Dialog,
    DialogContent,
    DialogContentText,
    DialogActions,
    RadioGroup,
    Radio,
    FormControlLabel,
    Divider,
} from '@material-ui/core'
import { isEqual } from 'lodash-es'

import { selectProjectById } from '../../redux/projects'
import { useSelector } from '../../redux'
import { ListItemButtonLink } from '../common/common'
import { fetchPlan, selectPlanExists, selectPlanStateById } from '../../redux/plans'
import { useHistory } from 'react-router-dom'
import * as api from '../../client'

interface PlanIterationValue {
    id: string
    details: api.IPlanDetails
}
// useForEachPlan builds an array by running the callback for each plan, sorted by date with newest first
function useForEachPlan(ids: string[]): (callback: (v: PlanIterationValue) => ReactElement) => ReactElement[] {
    const plans = useSelector(
        (state) =>
            ids
                .map((id) => {
                    const plan = selectPlanStateById(state, id)
                    const planExists = selectPlanExists(state, id)
                    return {
                        id,
                        planExists,
                        fetchErrMsg: plan?.fetchErrMsg,
                        detailsFetched: planExists && plan.workingCopy.detailsFetched,
                        details: plan?.workingCopy.details,
                    }
                })
                .sort((a, b) => {
                    if (!a.details) {
                        return -1
                    } else if (!b.details) {
                        return 1
                    } else {
                        const dateComp = b.details.date.localeCompare(a.details.date)
                        if (dateComp === 0) {
                            // make the sort stable even if two plans have the same date
                            return a.id.localeCompare(b.id)
                        } else {
                            return dateComp
                        }
                    }
                }),
        isEqual
    )
    const dispatch = useDispatch()

    // fetch any missing plans
    useEffect(() => {
        ids.forEach((id) => {
            dispatch(fetchPlan(id))
        })
    }, [ids, dispatch])

    return (callback: (v: PlanIterationValue) => ReactElement) =>
        plans
            .filter((v) => v.planExists && (v.fetchErrMsg || v.detailsFetched)) // don't render anything for plans that haven't been fetched
            .map((v, i) =>
                v.fetchErrMsg ? <p key={'err' + i}>Error fetching</p> : callback({ id: v.id, details: v.details })
            )
}

export function PlanList(props: { ids: string[] }) {
    const forEachPlan = useForEachPlan(props.ids)

    return (
        <List disablePadding>
            {forEachPlan((plan) => (
                <ListItemButtonLink key={plan.id} listItemProps={{ dense: true }} to={`/plan/${plan.id}`}>
                    <ListItemText
                        primary={plan.details.date}
                        primaryTypographyProps={{ variant: 'body2', color: 'primary' }}
                        secondary={plan.details.committed ? null : 'draft'}
                    />
                </ListItemButtonLink>
            ))}
        </List>
    )
}

// Calling cancel should lead to the dialogue no longer being rendered
export function NewPlanDialogue(props: { project: string; cancel: () => void }) {
    const history = useHistory()
    const [planId, setPlanId] = useState('new')

    const create = () => {
        const params = new URLSearchParams({ projects: props.project })
        if (planId !== 'new') {
            params.append('template', planId)
        }
        history.push(`plan/new?${params.toString()}`)
    }

    return (
        <Dialog
            open={true}
            onClose={props.cancel}
            aria-labelledby="alert-dialog-title"
            aria-describedby="alert-dialog-description"
        >
            <DialogContent>
                <DialogContentText id="alert-dialog-description">
                    Create new plan from scratch, or select an existing plan from this project to base it on.
                </DialogContentText>
                <PlanSelector project={props.project} select={setPlanId} selected={planId} />
            </DialogContent>
            <DialogActions>
                <Button onClick={props.cancel} color="primary">
                    Cancel
                </Button>
                <Button onClick={create} color="primary" autoFocus>
                    Create
                </Button>
            </DialogActions>
        </Dialog>
    )
}

function PlanSelector(props: { project: string; selected: string; select: (planId: string) => void }) {
    // It would be nice to be able to select a plan from any project, but for now you can just select ones from this project
    const plans = useSelector((state) => selectProjectById(state, props.project)).plans
    const forEachPlan = useForEachPlan(plans)

    return (
        <RadioGroup
            onChange={(event) => {
                props.select(event.target.value)
            }}
            value={props.selected}
        >
            <FormControlLabel value={'new'} key={'new'} control={<Radio />} label={'Start from a fresh plan'} />
            <Divider />
            {forEachPlan((plan) => (
                <FormControlLabel
                    value={plan.id}
                    key={plan.id}
                    control={<Radio />}
                    label={plan.details.date + (plan.details.committed ? '' : ' (draft)')}
                />
            ))}
        </RadioGroup>
    )
}
