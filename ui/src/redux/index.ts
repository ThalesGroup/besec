import { Action, configureStore } from '@reduxjs/toolkit'
import { TypedUseSelectorHook, useSelector as origUseSelector } from 'react-redux'
import { ThunkAction } from 'redux-thunk'

import practicesReducer, { PracticesState } from './practices'
import plansReducer, { PlansState } from './plans'
import projectsReducer, { ProjectsState } from './projects'
import notificationsReducer, { NotificationsState } from './notifications'
import sessionReducer, { SessionState } from './session'

export interface Store {
    practices: PracticesState
    plans: PlansState
    projects: ProjectsState
    notifications: NotificationsState
    session: SessionState
}

export const store = configureStore<Store>({
    reducer: {
        practices: practicesReducer,
        plans: plansReducer,
        projects: projectsReducer,
        notifications: notificationsReducer,
        session: sessionReducer,
    },
})

// Adapted from NonFunctionProperties in the TypeScript handbook: useful as client.ts has interface definitions
// that capture the types we want, but they often aren't serialiazable because they contain types
// that are classes. As all the autogenerated classes have an init function, we can filter them out.
interface HasInit {
    init(data?: any): void
}
type SerializablePropertyNames<T> = { [K in keyof T]: T[K] extends HasInit ? never : K }[keyof T]
export type SerializableProperties<T> = Pick<T, SerializablePropertyNames<T>>

// A typed version of useSelector, for use throughout the project
export const useSelector: TypedUseSelectorHook<Store> = origUseSelector

export type AppThunkAction = ThunkAction<void, Store, null, Action<string>>

function isErrorWithMessage(error: unknown): error is { message: string } {
    return (
        typeof error === 'object' &&
        error !== null &&
        'message' in error &&
        typeof (error as Record<string, unknown>).message === 'string'
    )
}

export function getErrorMessage(error: unknown) {
    if (isErrorWithMessage(error)) return error.message

    try {
        return JSON.stringify(error)
    } catch {
        // fallback in case there's an error stringifying
        // like with circular references for example.
        return String(error)
    }
}
