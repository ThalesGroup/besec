import React, { useEffect } from 'react'
import { initializeApp, getApps } from 'firebase/app'
import { getAuth, connectAuthEmulator, onAuthStateChanged, Auth, User } from 'firebase/auth'
import FirebaseAuth from 'react-firebaseui/FirebaseAuth'
import { useDispatch } from 'react-redux'
import {
    Button,
    Avatar,
    AvatarProps,
    Popper,
    MenuItem,
    MenuList,
    makeStyles,
    Fade,
    Paper,
    ClickAwayListener,
    ListItemText,
    Divider,
} from '@material-ui/core'
import { ArrowDropDown as ArrowDropDownIcon } from '@material-ui/icons'
import clsx from 'clsx'

import {
    initialize,
    login,
    logout,
    selectLoggedIn,
    selectUserInfo,
    selectAuthConfig,
    selectSigninOptions,
    selectInitialized,
} from '../redux/session'
import { useSelector } from '../redux'
import ApiClient from '../ApiClient'
import * as api from '../client'

// useFirebaseAuth initializes and returns an App-wide firebase auth instance
export function useFirebaseAuth() {
    const dispatch = useDispatch()
    const config = useSelector(selectAuthConfig)

    // we need to re-render on changes to these
    useSelector(selectInitialized)
    useSelector(selectLoggedIn)

    // Initialize app state
    // Listen to auth changes and propagate them to the store
    useEffect(() => {
        if (
            // only initialize once, no matter how many components use the hook.
            getApps().length > 0 ||
            // wait for auth configuration to be retrieved from the server
            !config
        ) {
            return
        }

        // do NOT return the unsubscribe function to cleanup - we need this callback
        // to trigger even if the original component is gone, as we only register one callback globally
        initializeApp({ apiKey: config.gcpPublicApiKey, authDomain: config.gcpAuthDomain })
        const auth = getAuth()
        if (config.emulatorUrl) {
            connectAuthEmulator(auth, config.emulatorUrl)
        }

        onAuthStateChanged(auth, async (user: User | null) => {
            if (user) {
                const { uid, providerId, phoneNumber } = user
                let { email, displayName, photoURL } = user

                // If this was a federated sign-in, most of these attributes won't be populated.
                // Instead we need to fetch them out of the ID Token
                const idToken = await user.getIdTokenResult()
                if (idToken.signInProvider && config.providers.hasOwnProperty(idToken.signInProvider)) {
                    // We have a mapping from this SAML provider's claims to the standard user attributes
                    const attrs = (idToken.claims?.['firebase'] as any)?.['sign_in_attributes']
                    if (attrs) {
                        const provider = config.providers[idToken.signInProvider]
                        email = attrs[provider.email]
                        if (provider.name) {
                            displayName = attrs[provider.name]
                        } else {
                            displayName = attrs[provider.firstName!] + ' ' + attrs[provider.surname!]
                        }
                        if (provider.pictureURL) {
                            photoURL = attrs[provider.pictureURL]
                        }
                    } else {
                        console.warn('Firebase sign_in_attributes missing from idToken of a configured SAML provider')
                    }
                }

                ApiClient.client.getIdToken = () => user.getIdToken()
                ApiClient.client.loggedIn().then(
                    // we don't care what the response is - just need to hit it
                    () => undefined,
                    () => undefined
                )
                dispatch(login({ uid, email, displayName, providerId, phoneNumber, photoURL }))
            } else {
                ApiClient.client.getIdToken = null
                dispatch(logout())
            }
        })
        dispatch(initialize())
    }, [config, dispatch])

    return getApps().length > 0 ? getAuth() : undefined
}

export function Login(props: { auth?: Auth }) {
    const signInOptions = useSelector(selectSigninOptions)

    if (!props.auth || signInOptions?.length === 0) {
        return <></>
    }

    const uiConfig = {
        // Popup signin flow rather than redirect flow. TODO: this would be better as redirect
        signInFlow: 'popup',
        // do nothing on login - useFirebaseAuth sets up an observer
        callbacks: {
            signInSuccessWithAuthResult: (authResult: any) => false,
        },
        signInOptions: signInOptions,
    }
    return <FirebaseAuth uiConfig={uiConfig} firebaseAuth={props.auth} />
}

// Returns the component if logged in, otherwise returns the Login component
export function Authenticated(loggedIn: boolean, component: React.ComponentType<any>, auth?: Auth) {
    if (loggedIn) {
        return component
    } else {
        const login = () => <Login auth={auth} />
        login.displayName = 'Login'
        return login
    }
}

// UserMenu component adapted from https://github.com/entaxy-project/entaxy (MIT)
const useStyles = makeStyles((theme) => ({
    root: {
        display: 'flex',
        marginLeft: 'auto',
    },
    popper: {
        zIndex: theme.zIndex.appBar + 1,
    },
    menuIcon: {
        marginRight: '5px',
        verticalAlign: 'bottom',
        fill: theme.palette.text.secondary,
    },
    smallAvatar: { height: '1.4em', width: '1.4em', fontSize: '1.1em' },
    userSpan: { display: 'inline-flex', alignItems: 'center' },
}))

export function UserMenu() {
    const classes = useStyles()
    const loggedIn = useSelector(selectLoggedIn)
    const userInfo = useSelector(selectUserInfo)
    const auth = useFirebaseAuth()
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null)

    if (!(loggedIn && userInfo)) return null

    const handleToggle = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(anchorEl ? null : event.currentTarget)
    }

    const open = Boolean(anchorEl)
    const id = open ? 'simple-popper' : undefined

    return (
        <div className={classes.root}>
            <Button color="inherit" aria-describedby={id} onClick={handleToggle} data-testid="userNavButton">
                <InitialsAvatar
                    url={userInfo.photoURL}
                    fullName={userInfo.displayName}
                    alt={userInfo.displayName ? userInfo.displayName : undefined}
                />
                <ArrowDropDownIcon />
            </Button>
            <Popper id={id} open={open} anchorEl={anchorEl} transition className={classes.popper}>
                {({ TransitionProps }) => (
                    <Fade {...TransitionProps} timeout={300}>
                        <Paper>
                            <ClickAwayListener onClickAway={() => setAnchorEl(null)}>
                                <MenuList role="menu">
                                    <MenuItem disabled={true}>
                                        <ListItemText>{userInfo.displayName}</ListItemText>
                                    </MenuItem>
                                    <Divider />
                                    <MenuItem onClick={() => auth?.signOut()} data-testid="logoutButton">
                                        <ListItemText primary="Sign out" />
                                    </MenuItem>
                                </MenuList>
                            </ClickAwayListener>
                        </Paper>
                    </Fade>
                )}
            </Popper>
        </div>
    )
}

type InitialsAvatarProps = {
    url?: string | null
    fullName?: string | null
} & AvatarProps
function InitialsAvatar({ url, fullName, ...rest }: InitialsAvatarProps) {
    if (url) {
        return <Avatar src={url} {...rest} />
    } else {
        let text = '?'
        if (fullName) {
            text = fullName
                .split(' ')
                .map((word) => word.charAt(0))
                .join('')
        }
        return <Avatar {...rest}>{text}</Avatar>
    }
}

export function UserNameIcon(props: { user?: api.IAuthor; className?: string }) {
    const classes = useStyles()
    return (
        <span className={clsx(props.className, classes.userSpan)}>
            <InitialsAvatar
                url={props.user?.pictureUrl}
                fullName={props.user?.name}
                classes={{ img: classes.smallAvatar }}
            />
            {props.user?.name ?? '?'}
        </span>
    )
}
