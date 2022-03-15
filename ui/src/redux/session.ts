import { createSlice, PayloadAction, createSelector, createAsyncThunk } from '@reduxjs/toolkit'
import { UserInfo } from 'firebase/auth'
import { auth as firebaseAuth } from 'firebaseui'

import { Store } from '.'
import * as api from '../client'
import ApiClient from '../ApiClient'

export interface SessionState {
    initialized: boolean
    loggedIn: boolean
    info?: UserInfo
    authConfigIsFetching: boolean
    authConfigFetched: boolean
    authConfig?: api.IAuthConfig
}

const sessionSlice = createSlice({
    name: 'session',
    initialState: {
        loggedIn: false,
    } as SessionState,
    reducers: {
        initialize: (state) => {
            state.initialized = true
        },
        login: (state, action: PayloadAction<UserInfo>) => {
            state.loggedIn = true
            state.info = action.payload
        },
        logout: (state) => {
            state.loggedIn = false
        },
    },
    extraReducers: (builder) => {
        // The `builder` callback form is used here because it provides correctly typed reducers from createAsyncThunk
        builder
            .addCase(fetchAuthConfig.fulfilled, (state, { payload }) => {
                state.authConfigIsFetching = false
                state.authConfig = payload
            })
            .addCase(fetchAuthConfig.pending, (state) => {
                state.authConfigIsFetching = true
            })
            .addCase(fetchAuthConfig.rejected, (state) => {
                state.authConfigIsFetching = false
            })
    },
})

export const fetchAuthConfig = createAsyncThunk<api.IAuthConfig, void, { state: Store }>(
    'session/fetchAuthConfig',
    async () => {
        const fetchedConfig = await ApiClient.client.getAuthConfig()
        return fetchedConfig.toJSON()
    },
    {
        condition: (_, { getState }) => {
            const { authConfig, authConfigIsFetching } = getState().session
            if (authConfig || authConfigIsFetching) {
                // Already fetched or in progress, don't need to re-fetch
                return false
            }
        },
    }
)

const selectSessionState = (store: Store) => store.session
export const selectInitialized = createSelector(selectSessionState, (session) => session.initialized)
export const selectLoggedIn = createSelector(selectSessionState, (session) => session.loggedIn)
export const selectUserInfo = createSelector(selectSessionState, (session) => session.info)

const selectAuthConfigState = createSelector(selectSessionState, (session) => session.authConfig)
const selectSamlProviders = createSelector(selectAuthConfigState, (config) => {
    return config?.providers.reduce(
        (providerMap, provider) =>
            provider.samlClaims ? { [provider.id]: provider.samlClaims, ...providerMap } : providerMap,
        {} as { [id: string]: api.SamlProviderClaimsMap }
    )
})
export const selectAuthConfig = createSelector(selectSessionState, selectSamlProviders, (session, providers) => {
    if (!session.authConfig) return undefined
    return {
        ...session.authConfig,
        providers: providers!,
    }
})

export const selectSigninOptions = createSelector(
    selectAuthConfigState,
    (config) =>
        config?.providers.map((p) => ({ provider: p.id, ...p.signInOptions })) as firebaseAuth.Config['signInOptions']
)

export const { initialize, login, logout } = sessionSlice.actions

export default sessionSlice.reducer
