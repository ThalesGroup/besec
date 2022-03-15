import { createSlice, createSelector, PayloadAction } from '@reduxjs/toolkit'
import { Dispatch } from 'redux'
import moment from 'moment'

import { Store, AppThunkAction, getErrorMessage } from '.'
import * as api from '../client'
import ApiClient from '../ApiClient'
import { fetchProjects } from './projects'
import { notify } from './notifications'
import _ from 'lodash'
import {
    selectLatestPracticesVersion,
    selectPractices,
    selectPracticesVersionFromPractices,
    selectPracticesVersionState,
} from './practices'

// a special plan ID for the unique unsaved plan, if it exists
// As plan IDs are generated server-side, this won't conflict
export const NEWPLAN = 'new'

export interface PlansState {
    ids: string[]
    deleted: string[] // a record of plan IDs that have been successfully deleted
    byId: Record<string, PlanState>
    creatingNewPlan: boolean // whether the new plan is currently being initialized from the practice definitions
}

interface PlanState {
    id: string

    // Whether the workingCopy plan is identical to the latest revision
    saved: boolean
    isSaving: boolean
    saveErrMsg?: string

    isDeleting: boolean
    deleteErrMsg?: string

    // The plan revisionIds and versions can be retrieved separately
    historyFetched: boolean
    fetchingHistory: boolean

    // Captures an error when retrieving the plan itself, not any particular revision of it
    fetchErrMsg?: string

    // a sorted list of ids, from oldest to newest. Only populated if history is fetched.
    revisionIds: string[]
    // the latest revision ID. If historyFetched, equal to the last element of revisionIds, otherwise this is the only record of the latest revision ID.
    latestRevisionId: string
    // read-only: these always reflect what is persisted in the database
    // Old revisions cannot be fetched without also fetching the plan history.
    revisionsById: Record<string, PlanRevision>
    // version info keyed on revision ID
    versions: Record<string, PlanVersion>
    // Can be edited. If no edits have been made, it holds the same data as revisionsById[latestRevisionId]
    workingCopy: PlanRevision
}

export interface PlanVersion {
    author: api.IAuthor
    time: string // serialized moment
}

export interface PlanRevision {
    // Plan details are always fetched, but answers can be retrieved independently
    // Note there is only a single fetch error message, on the parent plan
    detailsFetched: boolean
    fetchingDetails: boolean
    responsesReady: boolean // answers ready indicates they are either fetched, or it's the new plan and they've been initialized
    fetchingResponses: boolean
    fetchErrMsg?: string

    details: api.IPlanDetails
    practicesVersion: string
    answersById?: AnswerRecord
    responsesById?: TaskRecord
}

// Answers keyed by practice then question ID
type AnswerRecord = Record<string /* practice */, Record<string /* question ID */, api.IAnswer>>
// Task responses, excluding answers, keyed by practice then task ID
type TaskRecord = Record<string /* practice */, Record<string /* task ID */, TaskResponseDetails>>
type TaskResponseDetails = {
    priority?: boolean
    issues?: string[]
    references?: string
}

type History = Pick<PlanState, 'revisionIds' | 'versions'>

const initialState: PlansState = {
    ids: [],
    deleted: [],
    byId: {},
    creatingNewPlan: false,
}

const emptyRevision: PlanRevision = {
    detailsFetched: false,
    fetchingDetails: false,
    responsesReady: false,
    fetchingResponses: false,

    details: {
        date: '',
        committed: false,
        projects: [],
        maturity: {},
    },
    practicesVersion: '',
    answersById: {},
    responsesById: {},
}
const emptyPlan: PlanState = {
    id: '',

    historyFetched: false,
    fetchingHistory: false,

    saved: false,
    isSaving: false,
    isDeleting: false,

    revisionIds: [],
    latestRevisionId: '',
    versions: {},
    revisionsById: {},
    workingCopy: emptyRevision,
}

const plansSlice = createSlice({
    name: 'plans',
    initialState,
    reducers: {
        request: (
            state,
            action: PayloadAction<{ id: string; revisionId?: string; responses: boolean; history: boolean }>
        ) => {
            const p = state.byId[action.payload.id]
            p.fetchingHistory = action.payload.history
            let revision
            if (action.payload.revisionId) {
                if (!p.revisionsById[action.payload.revisionId]) {
                    p.revisionsById[action.payload.revisionId] = JSON.parse(JSON.stringify(emptyRevision))
                }
                revision = p.revisionsById[action.payload.revisionId]
            } else {
                revision = p.workingCopy
            }
            revision.fetchingDetails = true
            revision.fetchingResponses = action.payload.responses
        },
        receiveErr: (
            state,
            action: PayloadAction<{
                id: string
                revisionId?: string
                responses: boolean
                history: boolean
                msg: string
            }>
        ) => {
            const p = state.byId[action.payload.id]
            if (!p) {
                console.error('receivePlanErr: the plan with ID ' + action.payload.id + " doesn't exist in the store")
                return
            }
            if (action.payload.revisionId) {
                const revision = p.revisionsById[action.payload.revisionId]
                revision.fetchingDetails = false
                if (action.payload.responses) {
                    revision.fetchingResponses = false
                }
                revision.fetchErrMsg = action.payload.msg
            } else {
                // failed prior to even fetching a revision ID
                p.fetchErrMsg = action.payload.msg

                p.workingCopy.fetchingDetails = false
                if (action.payload.responses) {
                    p.workingCopy.fetchingResponses = false
                }
            }
            if (action.payload.history) {
                p.fetchingHistory = false
            }
        },
        fetchedRevision: (
            state,
            action: PayloadAction<{
                planId: string
                revisionId: string
                details: PlanRevision['details']
                history?: History
                responses?: { version: string; answers: AnswerRecord; tasks: TaskRecord }
            }>
        ) => {
            const p = state.byId[action.payload.planId]
            if (!(action.payload.revisionId in p.revisionsById)) {
                p.revisionsById[action.payload.revisionId] = JSON.parse(JSON.stringify(emptyRevision))
            }
            p.fetchErrMsg = undefined
            const revision = p.revisionsById[action.payload.revisionId]

            revision.fetchingDetails = false
            revision.detailsFetched = true
            revision.details = action.payload.details

            if (action.payload.history) {
                p.fetchingHistory = false
                p.historyFetched = true
                p.versions = action.payload.history.versions
                p.revisionIds = action.payload.history.revisionIds
                p.latestRevisionId = action.payload.history.revisionIds[action.payload.history.revisionIds.length - 1]
            } else {
                if (p.latestRevisionId === '') {
                    // We've never fetched history, so this revision ID must be the latest
                    p.latestRevisionId = action.payload.revisionId
                }
            }
            if (action.payload.responses) {
                revision.fetchingResponses = false
                revision.responsesReady = true
                revision.answersById = action.payload.responses.answers
                revision.responsesById = action.payload.responses.tasks
                revision.practicesVersion = action.payload.responses.version
            }

            if (action.payload.revisionId === p.latestRevisionId) {
                p.saved = revisionsEqual(p.workingCopy, revision)
            }
        },
        updateWorkingCopy: (state, action: PayloadAction<{ planId: string; revId?: string }>) => {
            const p = state.byId[action.payload.planId]
            if (!p) {
                console.error("Tried to update working copy when the plan doesn't exist")
                return
            }
            const revId = action.payload.revId ?? p.latestRevisionId
            if (!(revId in p.revisionsById)) {
                // Nothing to do - we don't have the revision
                return
            }
            // Deep clone the revision. As we only have serializable objects in the store, this is a safe method.
            p.workingCopy = JSON.parse(JSON.stringify(p.revisionsById[revId]))
            p.saved = revId === p.latestRevisionId
        },
        setPlanState: (state, action: PayloadAction<{ id: string; plan: PlanState }>) => {
            state.byId[action.payload.id] = action.payload.plan
            if (!state.ids.includes(action.payload.id)) {
                state.ids.push(action.payload.id)
            }
            if (action.payload.id === NEWPLAN) {
                state.creatingNewPlan = false
            }
        },
        initPlanRevision: (state, action: PayloadAction<{ planId: string; revisionId: string }>) => {
            state.byId[action.payload.planId].revisionsById[action.payload.revisionId] = JSON.parse(
                JSON.stringify(emptyRevision)
            )
        },
        answerQuestion: (
            state,
            action: PayloadAction<{ id: string; pId: string; qId: string; value: api.IAnswer }>
        ) => {
            const p = state.byId[action.payload.id]
            if (p.isSaving) {
                console.warn('Ignored a question that was answered whilst saving')
            } else if (!p.workingCopy.answersById) {
                // If it's a new plan, this will have been initialized.
                // If it already exists, it doesn't make sense to answer questions prior to fetching the existing answers
                console.warn(
                    "Tried to answer a question when the questions for this plan haven't been fetched yet, ignoring"
                )
            } else {
                p.workingCopy.answersById[action.payload.pId][action.payload.qId] = action.payload.value

                // this is relatively expensive. If we do ` = p.saved && _.isEqual(action.payload.value, ...)` it's quick but doesn't allow the user to see if their current copy actually has any changes
                p.saved = revisionsEqual(p.workingCopy, p.revisionsById[p.latestRevisionId])
            }
        },
        respondToTask: (
            state,
            action: PayloadAction<{ id: string; pId: string; tId: string; response: TaskResponseDetails }>
        ) => {
            const p = state.byId[action.payload.id]
            if (p.isSaving) {
                console.warn('Ignored a response that was made whilst saving')
            } else if (!p.workingCopy.responsesById) {
                // If it's a new plan, this will have been initialized.
                // If it already exists, it doesn't make sense to answer questions prior to fetching the existing answers
                console.warn(
                    "Tried to add a task response when the responses for this plan haven't been fetched yet, ignoring"
                )
            } else {
                const resp = p.workingCopy.responsesById[action.payload.pId][action.payload.tId]
                Object.assign(resp, action.payload.response)
                if (resp.issues) {
                    // remove any duplicates and empty values
                    resp.issues = Array.from(new Set(resp.issues)).filter((e) => e !== '')
                }

                // this is relatively expensive. If we do ` = p.saved && _.isEqual(resp, ...)` it's quick but doesn't allow the user to see if their current copy actually has any changes
                p.saved = revisionsEqual(p.workingCopy, p.revisionsById[p.latestRevisionId])
            }
        },
        setDetail: (state, action: PayloadAction<{ id: string; revisionDetails: Partial<api.IPlanDetails> }>) => {
            const p = state.byId[action.payload.id]
            Object.assign(p.workingCopy.details, action.payload.revisionDetails)

            // this is relatively expensive. If we do ` = p.saved && ...` it's quick but doesn't allow the user to see if their current copy actually has any changes
            p.saved = revisionsEqual(p.workingCopy, p.revisionsById[p.latestRevisionId])
        },
        removeNewPlan: (state) => {
            state.ids = state.ids.filter((v) => v !== NEWPLAN)
            delete state.byId[NEWPLAN]
        },
        requestSave: (state, action: PayloadAction<{ id: string }>) => {
            state.byId[action.payload.id].isSaving = true
        },
        saved: (state, action: PayloadAction<{ id: string; response: api.IAnonymous }>) => {
            const p = state.byId[action.payload.id]
            p.isSaving = false
            p.saveErrMsg = ''
            p.revisionIds.push(action.payload.response.revisionId)

            // we don't have a definitive local copy - wait for it to be fetched
            p.revisionsById[action.payload.response.revisionId] = JSON.parse(JSON.stringify(emptyRevision))

            if (action.payload.id === NEWPLAN) {
                p.id = action.payload.response.planId
                p.saved = true // this is working around a bug in the previous p.saved assignment, where the working copy is not present?
            }
        },
        saveErr: (state, action: PayloadAction<{ id: string; msg: string }>) => {
            const p = state.byId[action.payload.id]
            p.isSaving = false
            // don't set p.saved - if it happens that workingCopy already matches the latest revision (perhaps due to another save), the fact we failed to save now doesn't matter
            p.saveErrMsg = action.payload.msg
        },

        requestDelete: (state, action: PayloadAction<{ id: string }>) => {
            state.byId[action.payload.id].isDeleting = true
        },
        deleted: (state, action: PayloadAction<string>) => {
            delete state.byId[action.payload]
            state.ids = state.ids.filter((id) => id !== action.payload)
            state.deleted.push(action.payload)
        },
        deleteErr: (state, action: PayloadAction<{ id: string; msg: string }>) => {
            const p = state.byId[action.payload.id]
            p.isDeleting = false
            p.deleteErrMsg = action.payload.msg
        },
    },
})

// Are the data in the two plan revisions equal?
// Ignores the ephemeral aspects like error messages, fetch states
function revisionsEqual(a: PlanRevision, b: PlanRevision) {
    if (!a || !b) {
        return false
    }
    return (
        _.isEqual(a.details, b.details) &&
        _.isEqual(a.responsesById, b.responsesById) &&
        _.isEqual(a.answersById, b.answersById)
    )
}

// fetchPlan requests the plan's revision details and optionally its full responses and revision history
// If no revision ID is specified, the latest revision is retrieved
// If a revision ID is specified, and history hasn't been fetched before, it will fetch history regardless of the value of fetchHistory
// If the plan is already fetching the requested info, or the plan has been deleted, it does nothing.
// If the requested info has already been fetched, or there was an error fetching, it does nothing unless the force parameter is set
// The workingCopy will be replaced if the fetched revision is the latest revision, unless preserveWorkingCopy is set.
export function fetchPlan(
    planId: string,
    options: {
        revisionId?: string
        fetchResponses?: boolean
        fetchHistory?: boolean
        force?: boolean
        preserveWorkingCopy?: boolean
    } = {}
): AppThunkAction {
    return async (dispatch, getState) => {
        const { revisionId, fetchResponses = false, force = false, preserveWorkingCopy = false } = options
        let { fetchHistory = false } = options

        if (planId === NEWPLAN) {
            // the new plan can't be fetched
            return
        }

        let state = getState()
        if (!state.plans.ids.includes(planId)) {
            if (state.plans.deleted.includes(planId)) {
                // don't fetch a deleted plan
                return
            }
            const initPlan: PlanState = JSON.parse(JSON.stringify(emptyPlan))
            initPlan.id = planId
            // the plan doesn't exist in the store yet, create it
            dispatch(plansSlice.actions.setPlanState({ id: planId, plan: initPlan }))
            state = getState()
        }

        let p = state.plans.byId[planId]
        let revision: PlanRevision
        if (revisionId) {
            if (!p.historyFetched) {
                fetchHistory = true
            }
            if (!(revisionId in p.revisionsById)) {
                dispatch(plansSlice.actions.initPlanRevision({ planId, revisionId }))
                p = getState().plans.byId[planId]
            }
            revision = p.revisionsById[revisionId]
        } else {
            revision = p.workingCopy
        }

        if (
            revision.fetchingDetails &&
            p.fetchingHistory === fetchHistory &&
            revision.fetchingResponses === fetchResponses
        ) {
            // we're already on it!
            return
        } else if (
            !force &&
            (p.fetchErrMsg ||
                (revision.detailsFetched &&
                    (p.historyFetched || !fetchHistory) &&
                    (revision.responsesReady || !fetchResponses)))
        ) {
            // No need to fetch - there was an error previously or we already have at least everything being asked for
            return
        } else {
            dispatch(
                plansSlice.actions.request({
                    id: planId,
                    revisionId,
                    responses: fetchResponses,
                    history: fetchHistory,
                })
            )
        }

        let fetchedRevisionId: string
        let details: api.PlanDetails
        let fetchedVersions: api.RevisionVersion[]
        let fetchedResponses: api.IPracticeResponses
        try {
            let fetches: Promise<any>[]
            if (revisionId) {
                fetches = [ApiClient.client.getPlanRevision(planId, revisionId)]
            } else {
                fetches = [ApiClient.client.getPlan(planId)]
            }
            if (fetchHistory) {
                fetches.push(ApiClient.client.getPlanVersions(planId))
            }
            const results = await Promise.all(fetches)
            if (revisionId) {
                fetchedRevisionId = revisionId
                details = results[0]
            } else {
                const getPlanResponse: api.Anonymous2 = results[0]
                fetchedRevisionId = getPlanResponse.latestRevision
                details = getPlanResponse.plan.attributes
            }
            if (fetchHistory) {
                fetchedVersions = results[1]
            }
            if (fetchResponses) {
                fetchedResponses = await ApiClient.client.getPlanRevisionPracticeResponses(planId, fetchedRevisionId)
            }
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(
                plansSlice.actions.receiveErr({
                    id: planId,
                    responses: fetchResponses,
                    history: fetchHistory,
                    msg: msg,
                })
            )
            dispatch(notify({ key: 'fetcherr' + planId, message: 'Error fetching plan: ' + msg, variant: 'error' }))
            return
        }

        let history
        if (fetchHistory) {
            history = unmarshalVersions(fetchedVersions!)
        }
        let responses
        if (fetchResponses) {
            responses = unmarshalResponses(fetchedResponses!)
        }
        dispatch(
            plansSlice.actions.fetchedRevision({
                planId,
                revisionId: fetchedRevisionId,
                details: details.toJSON(),
                history,
                responses,
            })
        )
        if (fetchedRevisionId === getState().plans.byId[planId].latestRevisionId && !preserveWorkingCopy) {
            dispatch(plansSlice.actions.updateWorkingCopy({ planId }))
        }
    }
}

// createNewPlan returns a thunk action to initialize the new plan based on the practices in the state, or the provided template plan ID
export function createNewPlan(date: string, projectIds?: string[] | null, templateId?: string): AppThunkAction {
    return (dispatch, getState) => {
        const state = getState()
        if (state.plans.creatingNewPlan) {
            // already on it - as this might get called multiple times, we don't want to repeat the work
            return
        }

        const answersById: AnswerRecord = {}
        const responsesById: TaskRecord = {}
        const practices = selectPracticesVersionState(state, selectLatestPracticesVersion(state))
        if (!practices) {
            console.warn('Tried to create a new plan before retrieving the latest practices version')
            return
        }
        for (const practiceId of practices.ids) {
            answersById[practiceId] = {}
            for (const qId of practices.byId[practiceId].allQuestionIds) {
                answersById[practiceId][qId] = new api.Answer({ answer: api.Answer2.Unanswered }).toJSON()
            }
            responsesById[practiceId] = {}
            for (const task of practices.byId[practiceId].tasks) {
                responsesById[practiceId][task.id] = {}
            }
        }
        const plan: PlanState = {
            id: NEWPLAN,

            historyFetched: false,
            fetchingHistory: false,

            saved: false,
            isSaving: false,
            isDeleting: false,

            revisionIds: [],
            latestRevisionId: '',
            versions: {},
            revisionsById: {},
            workingCopy: {
                detailsFetched: false,
                fetchingDetails: false,
                responsesReady: true,
                fetchingResponses: false,

                details: {
                    date,
                    committed: false,
                    projects: projectIds ? projectIds : [],
                    maturity: {},
                },
                practicesVersion: state.practices.latestVersion!,
                answersById,
                responsesById,
            },
        }

        if (templateId) {
            if (!state.plans.ids.includes(templateId)) {
                plan.fetchErrMsg = "Internal error: the provided templateId wasn't found in the store " + templateId
            } else {
                const template = state.plans.byId[templateId]
                if (
                    template.fetchErrMsg ||
                    !template.workingCopy.detailsFetched ||
                    !template.workingCopy.responsesReady
                ) {
                    plan.fetchErrMsg =
                        'Internal error: the provided template is missing information necessary to use it as a template. ' +
                        'id: ' +
                        templateId +
                        ' fetch error: ' +
                        template.fetchErrMsg +
                        ' fetched: ' +
                        template.workingCopy.detailsFetched +
                        ' answers ready: ' +
                        template.workingCopy.responsesReady
                } else {
                    // Only copy the notes from the details - committed, date, and projects shouldn't be copied
                    plan.workingCopy.details.notes = template.workingCopy.details.notes

                    const templateAnswers = template.workingCopy.answersById!
                    const freshAnswers = plan.workingCopy.answersById!
                    Object.keys(freshAnswers).forEach((id) => {
                        // only replace answers where there is a matching ID, as the practices might have changed since the template plan was created
                        if (id in templateAnswers) {
                            // A question in a template could have an N/A answer, but then the practice changed to not allow N/A for that question.
                            // It's a pain to find a question given its ID with the current practice layout, so we're ignoring it.
                            // It will lead to a hopefully helpful error on save, that a user should be able to rectify.
                            Object.assign(freshAnswers[id], templateAnswers[id])
                        }
                    })

                    // repeat for task responses
                    const templateResponses = template.workingCopy.responsesById!
                    const freshResponses = plan.workingCopy.responsesById!
                    Object.keys(freshResponses).forEach((id) => {
                        // only replace answers where there is a matching ID, as the practices might have changed since the template plan was created
                        if (id in templateResponses) {
                            Object.assign(freshResponses[id], templateResponses[id])
                        }
                    })
                }
            }
        }

        dispatch(plansSlice.actions.setPlanState({ id: NEWPLAN, plan }))
    }
}

function unmarshalVersions(versions: api.RevisionVersion[]): History {
    const history: History = { revisionIds: [], versions: {} }
    for (const v of versions) {
        history['revisionIds'].push(v.revId)
        history['versions'][v.revId] = v.version.toJSON()
    }
    return history
}

// unmarshalResponses creates the store version of the answers and task responses
function unmarshalResponses(responses: api.IPracticeResponses): {
    version: string
    answers: AnswerRecord
    tasks: TaskRecord
} {
    const asById: AnswerRecord = {}
    const tsById: TaskRecord = {}
    for (const practiceId in responses.practiceResponses) {
        asById[practiceId] = {}
        const response = responses.practiceResponses[practiceId]
        if (response.practice) {
            for (const qId in response.practice) {
                asById[practiceId][qId] = response.practice[qId].toJSON()
            }
        }
        tsById[practiceId] = {}
        for (const tId in response.tasks) {
            const { answers, ...rest } = response.tasks[tId]
            for (const qId in response.tasks[tId].answers) {
                asById[practiceId][qId] = answers[qId].toJSON()
            }
            tsById[practiceId][tId] = rest
        }
    }
    return { version: responses.practicesVersion, answers: asById, tasks: tsById }
}

// marshalPlan creates a PracticeResponses instance out of the normalized answers in the store
function marshalRevision(practices: Record<string, api.IPractice>, rev: PlanState['workingCopy']): api.Body {
    if (!rev.answersById || !rev.responsesById) {
        throw new Error('Tried to marshal a revision but it has no answers')
    }
    const practiceResponses: { [practiceId: string]: api.PracticeResponse } = {}
    for (const practiceId in practices) {
        const practice = practices[practiceId]
        const practiceResp = new api.PracticeResponse()
        if (practice.questions) {
            practiceResp.practice = {}
            for (const q of practice.questions) {
                practiceResp.practice[q.id!] = new api.Answer(rev.answersById[practiceId][q.id!])
            }
        }
        for (const t of practice.tasks) {
            const qs: { [qID: string]: api.Answer } = {}
            for (const q of t.questions) {
                qs[q.id!] = new api.Answer(rev.answersById[practiceId][q.id!])
            }
            practiceResp.tasks[t.id] = new api.TaskResponse({ answers: qs, ...rev.responsesById[practiceId][t.id] })
        }
        practiceResponses[practiceId] = practiceResp
    }

    return api.Body.fromJS({
        details: api.PlanDetails.fromJS(rev.details),
        responses: { practicesVersion: rev.practicesVersion, practiceResponses },
    } as api.IBody)
}

// savePlan returns a thunk action to save the plan
// If this is the new plan, a fetch for the now-saved plan will be triggered and its ID will be changed,
// but it will remain available at byId[NEWPLAN] until removeNewPlan is dispatched
export function savePlan(id: string): AppThunkAction {
    return async (dispatch, getState) => {
        const state = getState()
        const p = state.plans.byId[id]
        if (p.isSaving) {
            return
        }
        const practiceVersion = selectPracticesVersionState(state, p.workingCopy.practicesVersion)
        if (!practiceVersion?.fetched) {
            console.warn(
                `Couldn't save revision as the practices at version ${p.workingCopy.practicesVersion} aren't present`
            )
            return
        }

        dispatch(plansSlice.actions.requestSave({ id }))

        const cachedWorkingCopy = JSON.parse(JSON.stringify(p.workingCopy))
        const revision = marshalRevision(practiceVersion.byId, cachedWorkingCopy)
        let ids: api.IAnonymous
        let response: string | api.Anonymous
        const err = (e: any) => {
            dispatch(plansSlice.actions.saveErr({ id, msg: e.message }))
            dispatch(notify({ key: 'saveerr' + id, message: 'Error saving plan: ' + e.message, variant: 'error' }))
        }
        if (id !== NEWPLAN) {
            // The plan exists, create a new revision
            try {
                response = await ApiClient.client.createPlanRevision(revision, id)
            } catch (e) {
                err(e)
                return
            }
            ids = { planId: id, revisionId: response }
        } else {
            // The plan doesn't exist, create it with its initial revision
            try {
                response = await ApiClient.client.createPlan(revision)
            } catch (e) {
                err(e)
                return
            }
            ids = response.toJSON()
            ;(dispatch as Dispatch<any>)(
                fetchPlan(ids.planId, { revisionId: ids.revisionId, fetchResponses: true, fetchHistory: true })
            )
        }

        dispatch(plansSlice.actions.saved({ id, response: ids }))
        // The server will have calculated metrics for the saved plan, as well as a new version entry, which we need to retrieve
        dispatch(fetchPlan(id, { revisionId: ids.revisionId, fetchResponses: true, fetchHistory: true, force: true }))
        dispatch(notify({ key: 'saved' + id, message: 'Saved', variant: 'success' }))
    }
}

// deletePlan returns a thunk action to save the plan
// You cannot delete the new plan - use removeNewPlan for that
export function deletePlan(id: string): AppThunkAction {
    return async (dispatch, getState) => {
        const state = getState()
        const p = state.plans.byId[id]
        if (p.isDeleting) {
            return
        }

        dispatch(plansSlice.actions.requestDelete({ id }))

        try {
            await ApiClient.client.deletePlan(id)
        } catch (e) {
            const msg = getErrorMessage(e)
            dispatch(plansSlice.actions.deleteErr({ id, msg: msg }))
            dispatch(notify({ key: 'deleteerr' + id, message: 'Error deleting plan: ' + msg, variant: 'error' }))
            return
        }

        // the projects are now stale - we could manually fix up any projects here, but it's easier to just re-request them
        dispatch(fetchProjects(true))
        dispatch(plansSlice.actions.deleted(id))
        dispatch(notify({ key: 'deleted' + id, message: 'Deleted plan', variant: 'success' }))
    }
}

export const selectPlanStateById = (state: Store, id: string) => state.plans.byId[id]
// Returns a mapping from plan ids to PlanStates
export const selectPlanStatesById = (state: Store, ids: string[]) => _.pick(state.plans.byId, ids)

const selectPlanAnswers = (state: Store, planId: string, revisionId?: string) => {
    const revision = selectPlanRevision(state, planId, revisionId)
    if (revision?.responsesReady) {
        return { answersById: revision.answersById, practicesVersion: revision.practicesVersion }
    }
    return undefined
}

// Returns the currentRevision of a plan if no revision ID is specified, otherwise that revision
export const selectPlanRevision = (state: Store, planId: string, revisionId?: string) => {
    if (revisionId) {
        return state.plans.byId[planId]?.revisionsById[revisionId]
    }
    return state.plans.byId[planId]?.workingCopy
}

export type PlanVersionID = PlanVersion & { revisionId: string }
export const selectPlanVersions = (state: Store, planId: string) =>
    state.plans.byId[planId]?.revisionIds.map(
        (revisionId) =>
            ({
                revisionId,
                ...state.plans.byId[planId].versions[revisionId],
            } as PlanVersionID)
    )

// Use with useMemo to create a selector that returns an array of practice IDs that have un-answered practice level questions
export const makeSelectUnansweredPracticeQuestions = () =>
    createSelector(selectPractices, selectPlanAnswers, (practices, planAnswers) => {
        if (!planAnswers?.answersById) {
            return []
        }
        return selectPracticesVersionFromPractices(practices, planAnswers.practicesVersion).reduce((unanswered, p) => {
            if (
                p.practice.questions?.some(
                    (q) => planAnswers.answersById![p.id][q.id!].answer === api.Answer2.Unanswered
                )
            ) {
                unanswered.push(p.id)
            }
            return unanswered
        }, [] as string[])
    })

export const makeSelectPlanDetailsValid = () =>
    createSelector(
        selectPlanRevision,
        (revision) =>
            revision &&
            revision.details.projects.length > 0 &&
            moment(revision.details.date, 'YYYY-M-D', true).isValid()
    )

// Create a selector that returns an array of practice ids that are applicable for the provided plan
// Practices with unanswered practice-level questions are NOT included - use makeSelectUnansweredPracticeQuestions for that
// Practices with N/A as an answer to any practice-level question are NOT included, per the schema definition
export const makeSelectApplicablePracticeIDs = () =>
    createSelector(selectPractices, selectPlanAnswers, (practices, planAnswers) => {
        if (!planAnswers?.answersById) {
            return []
        }
        return selectPracticesVersionFromPractices(practices, planAnswers.practicesVersion).reduce((applicable, p) => {
            let cond = p.practice.condition
            const qs = p.practice.questions
            if (qs && qs.length > 0 && cond) {
                let answered = true
                for (const q of qs) {
                    const answer = planAnswers.answersById![p.id][q.id!].answer
                    if (answer === api.Answer2.Unanswered || answer === api.Answer2.N_A) {
                        answered = false
                        break
                    }
                    cond = cond.replace(q.id!, answer === api.Answer2.Yes ? 'true' : 'false')
                }
                // evaluate the expression if fully answered. This is safe as the condition comes from the practice definition, not user input.
                // eslint-disable-next-line no-eval
                if (answered && eval(cond)) {
                    return applicable.concat([p.id])
                }
                return applicable
            }
            // no questions, it's always in scope
            return applicable.concat([p.id])
        }, [] as string[])
    })

// Create a selector that returns, for the provided plan ID, a list of practice IDs with unanswered task questions
export const makeSelectUnansweredPractices = () =>
    createSelector(selectPractices, selectPlanAnswers, (practices, planAnswers) => {
        if (!planAnswers?.answersById) {
            return []
        }
        return selectPracticesVersionFromPractices(practices, planAnswers.practicesVersion).reduce((unanswered, p) => {
            if (
                p.practice.tasks.some((task) =>
                    task.questions.some((q) => planAnswers.answersById![p.id][q.id!].answer === api.Answer2.Unanswered)
                )
            ) {
                return unanswered.concat([p.id])
            }
            return unanswered
        }, [] as string[])
    })

export const selectIsPlanReadonly = (store: Store, planId: string, revisionId?: string) => {
    if (revisionId) {
        // you can't change the past (or even the latest revision if you've specifically requested to view that)
        return true
    }
    return store.plans.byId[planId]?.workingCopy.details.committed ?? true
}

export const selectTaskResponses = (
    state: Store,
    planId: string,
    practiceId: string,
    taskId: string,
    revisionId?: string
) => {
    const revision = selectPlanRevision(state, planId, revisionId)
    return revision?.responsesById?.[practiceId]?.[taskId]
}

export const selectAnswersIfReady = (state: Store, planId: string) => {
    const plan = state.plans.byId[planId]
    if (plan?.workingCopy.responsesReady) {
        return (({ answersById, practicesVersion }) => ({ answersById, practicesVersion }))(plan.workingCopy)
    }
}

export const selectPlanDeleted = (state: Store, planId: string) => state.plans.deleted.includes(planId)
export const selectPlanExists = (state: Store, planId: string) => state.plans.ids.includes(planId)

export const { answerQuestion, setDetail, removeNewPlan, respondToTask, updateWorkingCopy } = plansSlice.actions
export default plansSlice.reducer
