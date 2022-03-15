import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { VariantType } from 'notistack'

import { Store } from '.'

export type NotificationsState = ReadonlyArray<
    Readonly<{ message: string; dismissed?: boolean; key: string; variant?: VariantType }>
>

const notificationsSlice = createSlice({
    name: 'notifications',
    initialState: [] as NotificationsState,
    reducers: {
        notify: (state, action: PayloadAction<{ message: string; key: string; variant?: VariantType }>) => {
            action.payload.key += Date.now()
            state.push(action.payload)
        },
        dismiss: (state, action: PayloadAction<string>) => {
            return state.map(n => (n.key === action.payload ? { ...n, dismissed: true } : n))
        },
        remove: (state, action: PayloadAction<string>) => {
            return state.filter(n => n.key !== action.payload)
        }
    }
})

export const selectNotifications = (state: Store) => state.notifications

export const { notify, dismiss, remove } = notificationsSlice.actions
export default notificationsSlice.reducer
