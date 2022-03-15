import React from 'react'
import { makeStyles, Typography, Paper, List, ListItemText, Theme, createStyles } from '@material-ui/core'
import { ListItemButtonLink, toolbarRelativeProperties } from './common'
import clsx from 'clsx'

const useStyles = makeStyles((theme: Theme) => {
    return createStyles({
        toc: {
            marginLeft: theme.spacing(2),
            position: 'sticky',
            '& ul': { 'padding-left': 0 }, // for some reason the list gets 40px of padding-left, don't know where it's coming from
            ...toolbarRelativeProperties('top')(theme),
            overflowY: 'auto',
            ...toolbarRelativeProperties('maxHeight', (tbHeight) => `calc(100vh - ${tbHeight}px)`)(theme),
        },
        currentSection: {
            fontWeight: theme.typography.fontWeightBold as number,
        },
        todo: {
            '&:after': { content: '" *"', color: 'red' },
        },
        notApplicable: {
            color: 'grey',
            '&:after': { content: '")"' },
            '&:before': { content: '"("' },
        },
    })
})

interface LinkType {
    text: string
    url: string
    current: boolean
    todo: boolean // whether there are questions to answer in this section before the plan can be committed
    applicable: boolean
}

export default function Toc(props: { links: LinkType[]; heading?: string }) {
    const classes = useStyles()

    return (
        <Paper className={classes.toc}>
            {props.heading && <Typography variant="subtitle1">{props.heading}</Typography>}
            <List>
                {props.links.map((link) => {
                    return (
                        <ListItemButtonLink key={link.url} to={link.url}>
                            <ListItemText
                                primary={link.text}
                                primaryTypographyProps={{
                                    className: clsx(
                                        link.current && classes.currentSection,
                                        link.todo && classes.todo,
                                        !link.todo && !link.applicable && classes.notApplicable
                                    ),
                                    variant: 'body2',
                                }}
                            />
                        </ListItemButtonLink>
                    )
                })}
            </List>
        </Paper>
    )
}
