import React from 'react'
import { Grid } from '@material-ui/core'

export function TwoColumns(props: { title?: JSX.Element; left: JSX.Element; right: JSX.Element }) {
    return (
        <>
            {props.title && props.title}
            <Grid container alignItems="baseline">
                <Grid item xs={6}>
                    {props.left}
                </Grid>
                <Grid item xs={6}>
                    {props.right}
                </Grid>
            </Grid>
        </>
    )
}
