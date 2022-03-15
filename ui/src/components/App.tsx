import React, { useEffect } from 'react'
import { Provider, useDispatch } from 'react-redux'
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom'

import './App.css'
import { store, useSelector } from '../redux'
import NavWrapper from './Navigation'
import Plan from './plan/Plan'
import Projects from './project/Projects'
import About from './About'
import Practices from './practice/Practices'
import Metrics from './Metrics'
import { fetchProjects } from '../redux/projects'
import { fetchPractices } from '../redux/practices'
import { selectLoggedIn, fetchAuthConfig } from '../redux/session'
import Notifier from './Notifier'
import { useFirebaseAuth, Authenticated } from './Auth'

function NotFound() {
    return <p>404: Nothing here!</p>
}

function Main() {
    const dispatch = useDispatch()

    const auth = useFirebaseAuth()
    const loggedIn = useSelector(selectLoggedIn)

    // Get login config
    useEffect(() => {
        dispatch(fetchAuthConfig())
    }, [dispatch])

    // preload data on login
    useEffect(() => {
        if (loggedIn) {
            dispatch(fetchProjects())
            dispatch(fetchPractices({ version: 'latest' }))
        }
    }, [loggedIn, dispatch])

    return (
        <>
            <NavWrapper>
                <Switch>
                    <Route path="/plans" component={Authenticated(loggedIn, Projects, auth)} />
                    <Route path="/plan/:planId/:section?" component={Authenticated(loggedIn, Plan, auth)} />
                    <Route path="/practices/:practice?" component={Authenticated(loggedIn, Practices, auth)} />
                    <Route path="/metrics/:projectName?" component={Authenticated(loggedIn, Metrics, auth)} />
                    <Route exact path="/" component={About} />
                    <Route component={NotFound} />
                </Switch>
            </NavWrapper>
            <Notifier />
        </>
    )
}

export default function App() {
    return (
        <Provider store={store}>
            <Router>
                <Main />
            </Router>
        </Provider>
    )
}
