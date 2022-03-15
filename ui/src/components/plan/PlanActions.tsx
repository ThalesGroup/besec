import React from 'react'
import { useDispatch } from 'react-redux'
import { Grid, createStyles, makeStyles, Theme, Button, Tooltip } from '@material-ui/core'
import { Edit as EditIcon, Send as SendIcon, Save as SaveIcon, Delete as DeleteIcon } from '@material-ui/icons'

import APIButton from '../common/APIButton'
import { savePlan, setDetail, removeNewPlan, deletePlan } from '../../redux/plans'
import { useHistory } from 'react-router-dom'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        planActionBar: { margin: theme.spacing(1, 0) },
        rightIcon: {
            marginLeft: theme.spacing(1),
            height: '1em',
        },
        deleteIcon: {
            height: '1em',
            color: 'red',
        },
        leftButton: {
            margin: theme.spacing(1),
            marginRight: 'auto',
        },
    })
)

interface PlanActionsProps {
    id: string
    new: boolean
    isDeleting: boolean
    isSaving: boolean
    dirty: boolean // the save button is disabled if this is false
    commitReady: boolean
    committed: boolean // ideally we'd be using the *server stored* committed value, as opposed to the version in the redux store.
}
export default function PlanActions(props: PlanActionsProps) {
    const dispatch = useDispatch()
    const classes = useStyles()
    const history = useHistory()

    let cancelOrDelete
    if (props.new) {
        cancelOrDelete = (
            <Tooltip title="Discard this unsaved plan" placement="top-start">
                <Button
                    onClick={() => {
                        history.push('/plans')
                        dispatch(removeNewPlan())
                    }}
                    className={classes.leftButton}
                    endIcon={<DeleteIcon />}
                >
                    {''}
                </Button>
            </Tooltip>
        )
    } else {
        cancelOrDelete = (
            <APIButton
                tooltipTitle="Permanently delete this draft plan"
                btnText=""
                confirmation={{ verb: 'Delete', question: 'Permanently delete this draft plan?' }}
                onClick={() => {
                    dispatch(deletePlan(props.id))
                }}
                inProgress={props.isDeleting}
                img={<DeleteIcon className={classes.deleteIcon} />}
                className={classes.leftButton}
            />
        )
    }

    let commitOrEdit
    if (props.committed) {
        commitOrEdit = (
            <APIButton
                tooltipTitle="Return this plan to a draft state to make further edits"
                btnText="Edit"
                onClick={() => {
                    dispatch(setDetail({ id: props.id, revisionDetails: { committed: false } }))
                    dispatch(savePlan(props.id))
                }}
                inProgress={props.isSaving}
                img={<EditIcon className={classes.rightIcon} />}
            />
        )
    } else {
        commitOrEdit = (
            <APIButton
                tooltipTitle="Finish this plan, recording it as the state of the project at this time. (You can revert this action.)"
                btnText="Commit"
                buttonProps={{ color: 'primary', variant: 'contained', disabled: !props.commitReady }}
                onClick={() => {
                    dispatch(setDetail({ id: props.id, revisionDetails: { committed: true } }))
                    dispatch(savePlan(props.id))
                }}
                inProgress={props.committed && props.isSaving}
                img={<SendIcon className={classes.rightIcon} />}
            />
        )
    }

    // a little progress bar of applicable tasks answered would be nice here:
    //   < CircularProgress className = { classes.progress } variant = "static" value = { 50} />
    return (
        <div className={classes.planActionBar}>
            <Grid id="actions" container justify="flex-end">
                {!props.committed && (
                    <>
                        {cancelOrDelete}
                        <APIButton
                            tooltipTitle={props.dirty ? 'Save a draft of this plan.' : 'No changes to save.'}
                            btnText="Save"
                            buttonProps={{ variant: 'contained', disabled: !props.dirty }}
                            onClick={() => {
                                dispatch(savePlan(props.id))
                            }}
                            inProgress={props.isSaving}
                            img={<SaveIcon className={classes.rightIcon} />}
                        />
                    </>
                )}
                {commitOrEdit}
            </Grid>
        </div>
    )
}
