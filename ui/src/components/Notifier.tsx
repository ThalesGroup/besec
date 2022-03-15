import React, { useState, useEffect } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { useSnackbar, OptionsObject, SnackbarProvider } from 'notistack'
import { IconButton } from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'

import { selectNotifications, dismiss, remove } from '../redux/notifications'


type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>

/**
 * Displays snackbar notifications queued in the redux store
 *
 * adapted from https://github.com/tlaziuk/pagepro
 * which in turn is a hooks adaptation of https://codesandbox.io/s/github/iamhosseindhv/notistack/tree/master/examples/redux-example
 */
function WrappedNotifier() {
    const [displayed, setDisplayed] = useState<ReadonlyArray<string | number>>([])
    const notifications = useSelector(selectNotifications)
    const dispatch = useDispatch()
    const { enqueueSnackbar, closeSnackbar } = useSnackbar()

    useEffect(() => {
        // Return a snackbar options object from some defaults and the provided notification-specific options
        const makeOptions: (options: Omit<OptionsObject, 'key'> & { key: string }) => OptionsObject = options => {
            const onClose = () => {
                dispatch(dismiss(options.key))
            }
            let autoHideDuration: number | null
            switch (options.variant) {
                case 'error':
                    autoHideDuration = null
                    break
                case 'success':
                    autoHideDuration = 2000
                    break
                default:
                    autoHideDuration = 5000
            }

            return {
                anchorOrigin: {
                    vertical: 'bottom',
                    horizontal: 'left'
                },
                ...options,
                autoHideDuration,

                action: (
                    <IconButton key="close" aria-label="close" color="inherit" onClick={onClose}>
                        <CloseIcon />
                    </IconButton>
                ),
                onClose
            }
        }

        // the displayed list grows until there are no notifications, then we reset it as we don't want to record historic notifications
        if (notifications.length === 0) {
            if (displayed.length > 0) {
                setDisplayed([])
            }
            return
        }

        for (const n of notifications) {
            if (n.dismissed) {
                closeSnackbar(n.key)
                dispatch(remove(n.key))
            }
        }

        const notDisplayed = notifications.filter(({ key }) => !displayed.includes(key))

        for (const { message, dismissed, ...rest } of notDisplayed) {
            enqueueSnackbar(message, makeOptions(rest))

            // Add to the displayed list
            setDisplayed(_ => [..._, rest.key])
        }
    }, [notifications, displayed, closeSnackbar, dispatch, enqueueSnackbar])

    return null
}

export default function Notifier() {
    return (
        <SnackbarProvider>
            <WrappedNotifier />
        </SnackbarProvider>
    )
}
