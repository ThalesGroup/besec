import { createAsyncThunk, createSelector, createSlice } from '@reduxjs/toolkit'

import { getErrorMessage, Store } from '.'
import * as api from '../client'
import ApiClient from '../ApiClient'

export interface ExtendedPractice extends api.IPractice {
    allQuestionIds: string[]
}
export interface PracticesState {
    versions: string[] // sorted. Every entry in versions exists in the database and has a corresponding entry in byVersion
    byVersion: Record<string, PracticesVersionState> // Entries in byVersion may not exist in the database (e.g. 'latest', and any requested version)
    latestVersion?: string
    fetchingHistory: boolean
    historyFetched: boolean
    historyFetchErr?: string
}
interface PracticesVersionState {
    isFetching: boolean
    fetched: boolean
    fetchErrMsg?: string
    ids: string[] // sorted by name (not ID)
    byId: Record<string, ExtendedPractice>
}

const practicesSlice = createSlice({
    name: 'practices',
    initialState: {
        versions: [],
        byVersion: {},
        fetchingHistory: false,
        historyFetched: false,
    } as PracticesState,
    reducers: {},
    extraReducers: (builder) => {
        // The `builder` callback form is used here because it provides correctly typed reducers from createAsyncThunk
        builder
            .addCase(fetchPractices.fulfilled, (state, { payload }) => {
                let version = payload.version
                if (version === 'latest') {
                    delete state.byVersion['latest'] // this is temporary whilst we're fetching the latest version
                    version = payload.receivedVersion
                    state.latestVersion = version
                }

                if (!state.versions.includes(version)) {
                    state.versions.push(version)
                }

                state.byVersion[version] = {
                    isFetching: false,
                    fetched: true,
                    fetchErrMsg: undefined,
                    ids: payload.ids,
                    byId: payload.byId,
                }
            })
            .addCase(fetchPractices.pending, (state, action) => {
                if (!(action.meta.arg.version in state.byVersion)) {
                    state.byVersion[action.meta.arg.version] = {
                        isFetching: true,
                        fetched: false,
                        ids: [],
                        byId: {},
                    }
                } else {
                    state.byVersion[action.meta.arg.version].isFetching = true
                }
            })
            .addCase(fetchPractices.rejected, (state, action) => {
                const v = state.byVersion[action.meta.arg.version]
                if (!v) {
                    console.error(
                        `Receive practices error for version '${action.meta.arg.version}' that haven't been requested!`
                    )
                    return
                }
                v.isFetching = false
                v.fetched = false
                v.fetchErrMsg = getErrorMessage(action.error)
            })

            .addCase(fetchPracticeHistory.fulfilled, (state, { payload }) => {
                state.versions = payload
                state.fetchingHistory = false
                state.historyFetchErr = undefined
                state.historyFetched = true
            })
            .addCase(fetchPracticeHistory.pending, (state) => {
                state.fetchingHistory = true
            })
            .addCase(fetchPracticeHistory.rejected, (state, action) => {
                state.fetchingHistory = false
                state.historyFetchErr = getErrorMessage(action.error)
            })
    },
})

// fetchPractices retrieves all the practices unless they've already been retrieved or had an error retrieving.
// If force is set, always retrieve them.
// could break this down to only fetch a practice when its needed, but if we want one we likely want them all
interface receivedPractice {
    version: string
    receivedVersion: string
    ids: PracticesVersionState['ids']
    byId: PracticesVersionState['byId']
}
export const fetchPractices = createAsyncThunk<
    receivedPractice,
    { version: string; force?: boolean },
    { state: Store }
>(
    'practices/fetchPractices',
    async ({ version = 'latest' }) => {
        const fetchedPractices = await ApiClient.client.getPractices(version)

        // Serialize the practices and add additional fields
        const byId: Record<string, ExtendedPractice> = {}
        for (const p in fetchedPractices.practices) {
            const practice = fetchedPractices.practices[p]
            const gatherQs = () => {
                const allQs: string[] = []
                const pushQs = (qs: api.Question[]) => {
                    for (const q of qs) {
                        allQs.push(q.id!) // the server guarantees all questions have IDs, even though the spec doesn't
                    }
                }
                if (practice.questions) {
                    pushQs(practice.questions)
                }
                for (const t of practice.tasks) {
                    pushQs(t.questions)
                }
                return allQs
            }
            byId[p] = { ...practice.toJSON(), allQuestionIds: gatherQs() }
        }

        const ids = Object.keys(fetchedPractices.practices)
        ids.sort((p1, p2) => p1.localeCompare(p2))

        return { version, receivedVersion: fetchedPractices.version, ids, byId }
    },
    {
        condition: ({ force = false, version = 'latest' }, { getState }) => {
            // only one request at a time, and only one ever unless forced
            const { latestVersion, byVersion } = getState().practices

            if (version === 'latest' && latestVersion && !force) {
                version = latestVersion
            }

            {
                const ms = byVersion[version]
                if (ms) {
                    if (ms.isFetching) {
                        // Nothing to do, already fetching this version
                        return false
                    }
                    // We assume the practices are static, so do nothing if we've already attempted to fetch
                    if (!force && (ms.fetched || ms.fetchErrMsg)) {
                        return false
                    }
                }
            }
        },
    }
)

// fetchPracticeHistory retrieves all of the practices version IDs
// If force is set, always retrieve them.
export const fetchPracticeHistory = createAsyncThunk<string[], boolean | undefined, { state: Store }>(
    'practices/fetchPracticeHistory',
    async () => {
        return await ApiClient.client.listPracticesVersions()
    },
    {
        condition: (force, { getState }) => {
            // only one request at a time, and only one ever unless forced
            const { fetchingHistory, historyFetched, historyFetchErr } = getState().practices
            if (fetchingHistory) {
                return false
            }
            return force || !(historyFetched || historyFetchErr)
        },
    }
)

export const selectPractices = (state: Store) => state.practices
export const selectPracticesVersionState = (state: Store, version?: string) => {
    if (version) {
        return state.practices.byVersion[version]
    }
    return undefined
}
// for use in other selectors
const selectPracticesVersionStateFromPractices = (state: PracticesState, version?: string) => {
    if (version) {
        return state.byVersion[version]
    }
    return undefined
}
const versionStateToList = (practices?: PracticesVersionState) =>
    practices?.ids.map((id) => ({ id, practice: practices.byId[id] })) ?? []
export const selectPracticesVersion = createSelector(selectPracticesVersionState, versionStateToList)
// for use in other selectors
export const selectPracticesVersionFromPractices = createSelector(
    selectPracticesVersionStateFromPractices,
    versionStateToList
)

export const selectPracticesByIdIfFetched = createSelector(selectPracticesVersionState, (practices) => {
    if (practices?.fetched) {
        return practices.byId
    }
    return {}
})

export const selectVersionOrLatest = (state: Store, version?: string) =>
    version ?? state.practices.latestVersion ?? 'latest'

export const selectLatestPracticesVersion = (state: Store) => state.practices.latestVersion

export const selectPracticeById = (state: Store, version?: string, id?: string) => {
    const pstate = selectPracticesVersionState(state, version)
    if (id && pstate?.ids.includes(id)) {
        return pstate.byId[id]
    }
    return undefined
}

export default practicesSlice.reducer
