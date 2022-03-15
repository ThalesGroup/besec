import { createSelector, createSlice, PayloadAction } from '@reduxjs/toolkit'

import * as api from '../client'
import ApiClient from '../ApiClient'
import { AppThunkAction, getErrorMessage, SerializableProperties, Store } from '.'
import { notify } from './notifications'

interface ProjectState {
    editing: boolean
    saved: boolean
    isSaving: boolean
    saveErrMsg?: string

    isDeleting: boolean
    deleteErrMsg?: string
    // no delete success display, as the entry is removed on success

    project: SerializedProject
    origAttributes?: api.IProjectDetails
}

export interface ProjectsState {
    fetched: boolean
    isFetching: boolean
    fetchErrMsg?: string

    ids: string[]
    byId: {
        [id: string]: ProjectState
    }

    // A single new project can be in progress at one time.
    newProject: {
        drafting: boolean // whether a new project is being created by the user at the moment
        valid: boolean // whether the current attributes meet the validation criteria to submit

        isCreating: boolean
        createErrMsg?: string
        createErrDisplay: boolean
        createSuccessDisplay: boolean

        attributes: api.IProjectDetails
    }
}
export type SerializedProject = SerializableProperties<api.IProject> & { attributes: api.IProjectDetails }

const initProject = (project: SerializedProject): ProjectsState['byId'][number] => ({
    editing: false,
    saved: true,
    isSaving: false,

    isDeleting: false,

    project,
})

const projectsSlice = createSlice({
    name: 'projects',

    initialState: {
        fetched: false,
        isFetching: false,

        ids: [],
        byId: {},
        newProject: {
            drafting: false,
            valid: false,
            isCreating: false,
            createErrDisplay: false,
            createSuccessDisplay: false,
            attributes: { name: '' },
        },
    } as ProjectsState,

    reducers: {
        requestFetchProjects: (state) => {
            state.isFetching = true
        },
        fetchedProjects: (state, action: PayloadAction<SerializedProject[]>) => {
            state.isFetching = false
            state.fetched = true
            state.fetchErrMsg = undefined
            state.ids = []
            state.byId = {}
            action.payload.forEach((p) => {
                state.byId[p.id] = initProject(p)
                state.ids.push(p.id)
            })
        },
        fetchProjectsErr: (state, action: PayloadAction<string>) => {
            state.isFetching = false
            state.fetched = false
            state.fetchErrMsg = action.payload
        },

        editProject: (state, action: PayloadAction<string>) => {
            state.byId[action.payload].editing = true
            state.byId[action.payload].origAttributes = state.byId[action.payload].project.attributes
        },
        cancelEditProject: (state, action: PayloadAction<string>) => {
            state.byId[action.payload].editing = false
            if (state.byId[action.payload].origAttributes) {
                state.byId[action.payload].project.attributes = state.byId[action.payload].origAttributes! // why is typescript not picking up the preceding test?
            }
        },
        setProjectAttributes: (state, action: PayloadAction<{ id: string; attributes: api.IProjectDetails }>) => {
            state.byId[action.payload.id].project.attributes = action.payload.attributes
        },
        requestSaveProject: (state, action: PayloadAction<string>) => {
            state.byId[action.payload].isSaving = true
        },
        savedProject: (state, action: PayloadAction<string>) => {
            const p = state.byId[action.payload]
            p.isSaving = false
            p.saved = true
            p.saveErrMsg = ''
            p.editing = false
        },
        saveProjectErr: (state, action: PayloadAction<{ id: string; msg: string }>) => {
            const p = state.byId[action.payload.id]
            p.isSaving = false
            p.saved = false
            p.saveErrMsg = action.payload.msg
        },

        requestDeleteProject: (state, action: PayloadAction<string>) => {
            state.byId[action.payload].isDeleting = true
        },
        deletedProject: (state, action: PayloadAction<string>) => {
            delete state.byId[action.payload]
            state.ids = state.ids.filter((id) => id !== action.payload)
        },
        deleteProjectErr: (state, action: PayloadAction<{ id: string; msg: string }>) => {
            const p = state.byId[action.payload.id]
            p.isDeleting = false
            p.deleteErrMsg = action.payload.msg
        },

        createNewProjectDraft: (state) => {
            state.newProject = {
                drafting: true,
                valid: false,
                isCreating: false,
                createErrDisplay: false,
                createSuccessDisplay: false,
                attributes: { name: '' },
            }
        },
        cancelNewProject: (state) => {
            state.newProject.drafting = false
        },
        setNewProjectAttributes: (state, action: PayloadAction<api.IProjectDetails>) => {
            state.newProject.attributes = action.payload
            state.newProject.valid = action.payload.name !== ''
        },
        requestCreateProject: (state) => {
            state.newProject.isCreating = true
        },
        createdProject: (state, action: PayloadAction<string>) => {
            state.ids.push(action.payload)
            state.byId[action.payload] = initProject({
                id: action.payload,
                plans: [],
                attributes: state.newProject.attributes,
            })
            state.newProject.drafting = false
            state.newProject.createSuccessDisplay = true
        },
        createProjectErr: (state, action: PayloadAction<string>) => {
            state.newProject.isCreating = false
            state.newProject.createErrDisplay = true
            state.newProject.createErrMsg = action.payload
        },
        createErrSeen: (state) => {
            state.newProject.createErrDisplay = false
        },
        createSuccessSeen: (state) => {
            state.newProject.createSuccessDisplay = false
        },
    },
})

// Retrieve all projects unless there was an error previously or they've already been retrieved.
// If force is set, always fetch them.
export function fetchProjects(force = false): AppThunkAction {
    return async (dispatch, getState) => {
        const s = getState().projects
        if (s.isFetching) {
            return
        }
        if (!force && (s.fetchErrMsg || s.fetched)) {
            return
        }

        dispatch(projectsSlice.actions.requestFetchProjects())
        let projects: api.Project[]
        try {
            projects = await ApiClient.client.listProjects()
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(projectsSlice.actions.fetchProjectsErr(msg))
            return
        }
        dispatch(projectsSlice.actions.fetchedProjects(projects.map((p) => p.toJSON())))
    }
}

export function createProject(): AppThunkAction {
    return async (dispatch, getState) => {
        dispatch(projectsSlice.actions.requestCreateProject())
        let id: string
        const details = api.ProjectDetails.fromJS(getState().projects.newProject.attributes)
        try {
            id = await ApiClient.client.createProject(details)
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(projectsSlice.actions.createProjectErr(msg))
            dispatch(
                notify({
                    key: 'createproject-maynotbeunique', // a second creation error whilst the first error is still visible will not show an error
                    message: 'Error creating project: ' + msg,
                    variant: 'error',
                })
            )
            return
        }
        dispatch(projectsSlice.actions.createdProject(id))
        dispatch(notify({ key: 'created' + id, message: 'Created new project', variant: 'success' }))
    }
}

export function saveProject(id: string): AppThunkAction {
    return async (dispatch, getState) => {
        dispatch(projectsSlice.actions.requestSaveProject(id))
        const details = api.ProjectDetails.fromJS(getState().projects.byId[id].project.attributes)
        try {
            await ApiClient.client.updateProject(details, id)
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(projectsSlice.actions.saveProjectErr({ id, msg: msg }))
            dispatch(notify({ key: 'saveerr' + id, message: 'Error saving project: ' + msg, variant: 'error' }))
            return
        }
        dispatch(projectsSlice.actions.savedProject(id))
        dispatch(notify({ key: 'saved' + id, message: 'Saved', variant: 'success' }))
    }
}

export function deleteProject(id: string): AppThunkAction {
    return async (dispatch) => {
        dispatch(projectsSlice.actions.requestDeleteProject(id))
        try {
            await ApiClient.client.deleteProject(id)
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(projectsSlice.actions.deleteProjectErr({ id, msg: msg }))
            dispatch(notify({ key: 'deleteerr' + id, message: 'Error deleting project: ' + msg, variant: 'error' }))
            return
        }
        dispatch(projectsSlice.actions.deletedProject(id))
    }
}

const selectProjects = (state: Store) => state.projects
export const selectProjectStateById = (state: Store, id: string) => state.projects.byId[id]
export const selectProjectById = createSelector(selectProjectStateById, (state) => state.project)
export const selectProjectsFetchStatus = createSelector(selectProjects, ({ fetched, isFetching, fetchErrMsg }) => ({
    fetched,
    isFetching,
    fetchErrMsg,
}))

export const selectProjectIdsSorted = createSelector(selectProjects, (projects) =>
    projects.ids
        .slice(0)
        .sort((a, b) =>
            projects.byId[a].project.attributes.name.localeCompare(projects.byId[b].project.attributes.name)
        )
)

export const selectNewProject = createSelector(selectProjects, (projects) => projects.newProject)

export const selectProjectNames = createSelector(selectProjects, (projects) =>
    projects.ids.map((id) => ({
        id,
        name: projects.byId[id].project.attributes.name,
    }))
)

export const selectProjectByName = (state: Store, name?: string) => {
    for (const pid of state.projects.ids) {
        if (state.projects.byId[pid].project.attributes.name === name) {
            return pid
        }
    }
    return undefined
}

export const {
    createNewProjectDraft,
    cancelNewProject,
    setNewProjectAttributes,
    editProject,
    cancelEditProject,
    setProjectAttributes,
} = projectsSlice.actions
export default projectsSlice.reducer
