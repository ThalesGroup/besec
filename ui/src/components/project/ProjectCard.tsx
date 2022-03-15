import React, { useState } from 'react'
import { useDispatch } from 'react-redux'
import {
    Typography,
    makeStyles,
    Theme,
    createStyles,
    Card,
    CardContent,
    CardActions,
    Button,
    TextField,
} from '@material-ui/core'
import { Delete as DeleteIcon } from '@material-ui/icons'
import { Action } from '@reduxjs/toolkit'

import Markdown from '../common/MaterialMarkdown'
import {
    setNewProjectAttributes,
    createProject,
    cancelNewProject,
    editProject,
    setProjectAttributes,
    saveProject,
    cancelEditProject,
    deleteProject,
    selectProjectStateById,
    selectNewProject,
} from '../../redux/projects'
import { useSelector } from '../../redux'
import APIButton from '../common/APIButton'
import { PlanList, NewPlanDialogue } from './ProjectPlans'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        card: {
            maxWidth: 345,
            margin: theme.spacing(2),
        },
        deleteIcon: {
            height: '1em',
            color: 'red',
        },
        leftButton: {
            marginRight: 'auto',
        },
    })
)

export default function ProjectCard(props: { create: boolean; id?: string }) {
    const { editing, isDeleting, isSaving, project } = useSelector((state) =>
        props.id
            ? selectProjectStateById(state, props.id)
            : { editing: false, isDeleting: false, isSaving: false, project: null }
    )
    const newProject = useSelector(selectNewProject)
    const dispatch = useDispatch()
    const [nameValue, setName] = useState('')
    const [descValue, setDesc] = useState('')
    const [showNewPlan, setShowNewPlan] = useState(false)
    const classes = useStyles()

    // This component has three major states: displaying, editing, and creating.
    // Editing and creating look identical, but have different data sources & actions.

    let cardContents: JSX.Element[], cardActions: JSX.Element[]
    const editCardContents = (setAttrs: () => void) => [
        <TextField
            key="name"
            id={'editProjectName'}
            label="Name"
            value={nameValue}
            onChange={(event) => setName(event.target.value)}
            onBlur={() => setAttrs()}
            onKeyPress={(ev) => {
                if (ev.key === 'Enter') {
                    setAttrs()
                }
            }}
            variant="outlined"
            fullWidth
            margin="normal"
        />,
        <TextField
            key="desc"
            id={'editProjectDesc'}
            label="Description"
            value={descValue}
            onChange={(event) => setDesc(event.target.value)}
            onBlur={() => setAttrs()}
            onKeyPress={(ev) => {
                if (ev.key === 'Enter') {
                    setAttrs()
                }
            }}
            variant="outlined"
            fullWidth
            multiline
            margin="normal"
        />,
    ]
    const editCardActions = (
        saveAction: any, // an action, but ThunkActions aren't Actions, annoyingly
        cancelAction: Action<string>,
        inProgress: boolean
    ) => [
        <Button
            key="cancel"
            size="small"
            color="primary"
            onClick={() => {
                dispatch(cancelAction)
            }}
            disabled={inProgress}
        >
            Cancel
        </Button>,
        <APIButton
            key="save"
            buttonProps={{ color: 'primary', size: 'small' }}
            onClick={() => {
                dispatch(saveAction)
            }}
            btnText="Save"
            inProgress={inProgress}
        />,
    ]

    if (project) {
        if (editing) {
            const setAttrs = () => {
                dispatch(
                    setProjectAttributes({ id: props.id!, attributes: { name: nameValue, description: descValue } })
                )
            }
            cardContents = editCardContents(setAttrs)
            const allowDelete = project.plans.length === 0
            cardActions = [
                <APIButton
                    key="delete"
                    buttonProps={{ size: 'small', disabled: !allowDelete }}
                    tooltipTitle={allowDelete ? undefined : 'Cannot delete a project that has plans associated with it'}
                    confirmation={{ verb: 'Delete', question: 'Delete this project?' }}
                    onClick={() => {
                        dispatch(deleteProject(props.id!))
                    }}
                    btnText=""
                    img={<DeleteIcon className={classes.deleteIcon} />}
                    className={classes.leftButton}
                    inProgress={isDeleting}
                />,
                ...editCardActions(saveProject(props.id!), cancelEditProject(props.id!), isSaving),
            ]
        } else {
            cardContents = [
                <Typography key="name" gutterBottom variant="h5" component="h2">
                    {project.attributes.name}
                </Typography>,

                <Markdown
                    key="desc"
                    source={project.attributes.description}
                    paraVariant="body2"
                    color="textSecondary"
                />,

                <PlanList key="list" ids={project.plans} />,
            ]
            cardActions = [
                <Button
                    key="edit"
                    size="small"
                    color="primary"
                    onClick={() => {
                        setName(project.attributes.name)
                        setDesc(project.attributes.description ?? '')
                        dispatch(editProject(project.id))
                    }}
                >
                    Edit Project
                </Button>,
                <Button key="new" size="small" color="primary" onClick={() => setShowNewPlan(true)}>
                    New Plan
                </Button>,
            ]
        }
    } else {
        if (!props.create) {
            throw new Error(
                'ProjectCard: either create must be true or a project ID must be provided for an existing project'
            )
        }

        const setAttrs = () => {
            dispatch(setNewProjectAttributes({ name: nameValue, description: descValue }))
        }

        cardContents = editCardContents(setAttrs)
        cardActions = editCardActions(createProject(), cancelNewProject(), newProject.isCreating)
    }

    return (
        <Card className={classes.card}>
            <CardContent>{cardContents}</CardContent>
            <CardActions>{cardActions}</CardActions>
            {showNewPlan && (
                <NewPlanDialogue
                    cancel={() => setShowNewPlan(false)}
                    project={props.id! /* Can't set showNewPlan from a new project card, so we'll always have an ID */}
                />
            )}
        </Card>
    )
}
