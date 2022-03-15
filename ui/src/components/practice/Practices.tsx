import React, { useEffect } from 'react'
import {
    Grid,
    Paper,
    CircularProgress,
    makeStyles,
    Theme,
    createStyles,
    Typography,
    NativeSelect,
} from '@material-ui/core'
import { useDispatch } from 'react-redux'
import { RouteComponentProps } from 'react-router-dom'
import { Looks4 } from '@material-ui/icons'

import { useSelector } from '../../redux'
import {
    fetchPractices,
    fetchPracticeHistory,
    selectVersionOrLatest,
    selectPracticesVersionState,
    selectPractices,
} from '../../redux/practices'
import Toc from '../common/Toc'
import Practice from './Practice'
import { ErrorMsg } from '../common/common'
import { Refresh } from '../common/Refresh'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        progress: { position: 'fixed', left: '50%', top: '25%' },
        paper: { padding: theme.spacing(2) },
        version: { padding: theme.spacing(1), marginBottom: theme.spacing(2), textAlign: 'right' },
        select: { minWidth: 130 },
    })
)

export default function Practices({ match, history, location }: RouteComponentProps<{ practice: string }>) {
    const classes = useStyles()

    const practice = match.params.practice
    const specifiedVersion = new URLSearchParams(location.search).get('version') ?? undefined

    const dispatch = useDispatch()
    const version = useSelector((state) => selectVersionOrLatest(state, specifiedVersion))
    const practiceState = useSelector((state) => selectPracticesVersionState(state, version))
    const { versions, fetchingHistory } = useSelector(selectPractices)

    useEffect(() => {
        dispatch(fetchPracticeHistory())
        if (!practiceState || (!practiceState.fetched && !practiceState.fetchErrMsg && !practiceState.isFetching)) {
            dispatch(fetchPractices({ version }))
        } // else we're currently fetching practices or there was an error fetching them, do nothing
    }, [practiceState, version, dispatch])

    let content
    if (practiceState?.fetchErrMsg) {
        content = <ErrorMsg message={`Error fetching practices (version ${version}): ${practiceState.fetchErrMsg}`} />
    } else if (!practiceState?.fetched) {
        content = <CircularProgress className={classes.progress} />
    } else if (practice) {
        if (practice in practiceState.byId) {
            content = <Practice practice={practiceState.byId[practice]} />
        } else {
            content = <ErrorMsg message={`Practice ${practice} not found!`} />
        }
    } else {
        content = (
            <>
                <Typography variant="body1" gutterBottom>
                    Please select a practice in the table of contents.
                </Typography>
                <br />
                <Typography variant="body1" gutterBottom>
                    Each practice is composed of a number of tasks, of increasing maturity levels. The maturity level of
                    a task is indicated with a number like this:{' '}
                    <Looks4 color="primary" style={{ verticalAlign: 'text-bottom' }} />. Maturity levels range between 1
                    and 4, and the goal is for all projects to reach level 4 maturity in each practice.
                </Typography>
                <Typography variant="body1" gutterBottom>
                    Each task also has one or more questions, which are used during planning to tell whether a project
                    should consider doing this task, or if it&apos;s already being done.
                </Typography>
                <Typography variant="body1" gutterBottom>
                    Not all tasks, or all practices, are applicable to every project. Practices that aren&apos;t
                    universally applicable have questions at the start to determine applicability to a project. When
                    doing planning, most task questions can be answered with &quot;N/A&quot;.
                </Typography>
            </>
        )
    }

    return (
        <Grid container direction="row-reverse">
            <Grid item sm={12} md={2} align-self="baseline">
                {practiceState?.ids && (
                    <Toc
                        links={practiceState.ids.map((practiceId) => ({
                            url: '/practices/' + practiceId + (specifiedVersion ? '?version=' + specifiedVersion : ''),
                            text: practiceState.byId[practiceId].name,
                            current: practiceId === practice,
                            todo: false,
                            applicable: true,
                        }))}
                    />
                )}
            </Grid>
            <Grid container direction="column" item sm={12} md={10}>
                <Paper className={classes.version}>
                    <em>Practices version: </em>
                    <NativeSelect
                        value={version}
                        onChange={(event) => {
                            history.push(`/practices/${practice ?? ''}?version=${event.target.value}`)
                        }}
                        className={classes.select}
                    >
                        {versions?.map((v) => (
                            <option key={v} value={v}>
                                {v}
                            </option>
                        ))}
                    </NativeSelect>
                    <Refresh
                        refreshing={practiceState?.isFetching || fetchingHistory}
                        refresh={() => {
                            dispatch(fetchPracticeHistory(true))
                            dispatch(fetchPractices({ version, force: true }))
                        }}
                        normalPositioning
                    />
                </Paper>
                <Paper className={classes.paper}>{content}</Paper>
            </Grid>
        </Grid>
    )
}
