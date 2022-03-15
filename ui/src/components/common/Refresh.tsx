import React, { useEffect, useState } from 'react'
import { Refresh as RefreshIcon } from '@material-ui/icons'
import { IconButton, makeStyles } from '@material-ui/core'
import clsx from 'clsx'

const useStyles = makeStyles(theme => ({
    root: { position: 'absolute', right: 0 },
    refreshing: { animation: '$spin 1s linear 0s infinite' },
    '@keyframes spin': {
        from: { transform: 'rotate(0deg)' },
        to: { transform: 'rotate(360deg)' }
    }
}))

// Unless normalPositioningg is set, Refresh is positioned absolutely, so requires a positioned parent element
export function Refresh(props: { refreshing: boolean; refresh: () => void; normalPositioning?: boolean }) {
    const classes = useStyles()
    const [spinning, setSpinning] = useState(props.refreshing)

    useEffect(() => {
        if (props.refreshing) {
            setSpinning(true)
        } // else, we will stop spinning once the animation has finished its current iteration
    }, [props.refreshing])

    return (
        <IconButton
            className={clsx(!props.normalPositioning && classes.root, spinning && classes.refreshing)}
            onClick={props.refresh}
            onAnimationIteration={() => {
                setSpinning(props.refreshing)
            }}
        >
            <RefreshIcon />
        </IconButton>
    )
}
