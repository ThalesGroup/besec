import React from 'react'
import { Link } from 'react-router-dom'
import { Typography } from '@material-ui/core'
import { useSelector } from 'react-redux'

import { Login, useFirebaseAuth } from './Auth'
import { selectLoggedIn } from '../redux/session'

export default function About() {
    const loggedIn = useSelector(selectLoggedIn)
    const auth = useFirebaseAuth()
    return (
        <>
            <Typography variant="body1" gutterBottom>
                BeSec is a practical framework of best practices, processes, and tools applied to product development
                activities to improve the overall security of our products. You can learn more about BeSec{' '}
                <a href="https://github.com/ThalesGroup/besec">here</a>.
            </Typography>

            <Typography variant="body1" gutterBottom>
                This tool allows teams to record their progress against a set of security maturity tasks across a number
                of <Link to={'/practices'}>practices</Link>, and <Link to={'/plans'}>plan</Link> what tasks to focus on
                next.
            </Typography>
            <Typography variant="body1" gutterBottom>
                <Link to={'/metrics'}>Metrics</Link> are gathered from these plans to allow the organization to review
                progress.
            </Typography>
            <Typography variant="body1" gutterBottom>
                Use the links in the side bar to navigate. You&apos;ll need an account to access the content - you
                should be able to log in with your corporate credentials.
            </Typography>
            <br />
            {!loggedIn && <Login auth={auth} />}
        </>
    )
}
