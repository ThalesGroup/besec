import React, { useEffect } from 'react'
import { Grid, CircularProgress, makeStyles, Theme, createStyles, Fab, Hidden } from '@material-ui/core'
import { Add as AddIcon } from '@material-ui/icons'
import { useDispatch } from 'react-redux'

import { useSelector } from '../../redux'
import {
    fetchProjects,
    createNewProjectDraft,
    selectProjectsFetchStatus,
    selectProjectIdsSorted,
    selectNewProject,
} from '../../redux/projects'
import { ErrorMsg } from '../common/common'
import { Refresh } from '../common/Refresh'

import ProjectCard from './ProjectCard'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        progress: { position: 'fixed', left: '50%', top: '25%' },
        fabextended: { margin: theme.spacing(0, 2) }, // to match the margin we put on the cards
        fab: { position: 'fixed', bottom: '1rem', right: '1rem' },
    })
)

export default function Projects() {
    const dispatch = useDispatch()
    const classes = useStyles()
    const { fetched, isFetching, fetchErrMsg } = useSelector(selectProjectsFetchStatus)
    const sortedIds = useSelector(selectProjectIdsSorted)
    const newProject = useSelector(selectNewProject)

    // Refresh on load whether or not we've already fetched
    useEffect(() => {
        dispatch(fetchProjects(true))
    }, [dispatch])

    let content
    if (!fetched) {
        if (fetchErrMsg) {
            content = <ErrorMsg message={fetchErrMsg} />
        } else {
            content = <CircularProgress className={classes.progress} />
        }
    } else {
        const newProjectCard = (
            <Grid item key={'newProject'}>
                <ProjectCard create={true} />
            </Grid>
        )
        const onAdd = () => {
            dispatch(createNewProjectDraft())
        }

        content = (
            <>
                <Hidden mdUp>
                    <Fab color="primary" aria-label="add" className={classes.fab} onClick={onAdd}>
                        <AddIcon />
                    </Fab>
                </Hidden>
                <Hidden smDown>
                    <Grid item xs={12}>
                        <Fab
                            variant="extended"
                            color="primary"
                            aria-label="add"
                            className={classes.fabextended}
                            onClick={onAdd}
                        >
                            <AddIcon />
                            Create Project
                        </Fab>
                    </Grid>
                </Hidden>

                {
                    // only render the new project card if we're drafting one
                    newProject.drafting && newProjectCard
                }

                {sortedIds.map((id) => (
                    <Grid item key={id}>
                        <ProjectCard create={false} id={id} />
                    </Grid>
                ))}
            </>
        )
    }

    return (
        <Grid container style={{ position: 'relative' }}>
            <Refresh
                refresh={() => {
                    dispatch(fetchProjects(true))
                }}
                refreshing={isFetching}
            />
            {content}
        </Grid>
    )
}
