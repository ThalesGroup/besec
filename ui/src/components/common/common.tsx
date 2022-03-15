import React from 'react'
import { Link as RouterLink, LinkProps as RouterLinkProps } from 'react-router-dom'
import { HashLink } from 'react-router-hash-link'
import { ListItem, Theme, makeStyles, Grid } from '@material-ui/core'
import { CSSProperties } from '@material-ui/styles'

const useStyles = makeStyles((theme) => ({
    error: {
        minHeight: 40,
        border: 'dashed red',
        borderWidth: 1,
        borderRadius: 3,
        padding: '0 30px',
        boxShadow: '0 3px 5px 2px rgba(255, 105, 135, .3)',
        margin: theme.spacing(2),
        whiteSpace: 'pre-wrap',
    },
}))

export function ErrorMsg(props: { message: string }) {
    const classes = useStyles()
    return (
        <Grid container justify="center">
            <Grid item className={classes.error}>
                <p>{props.message}</p>
            </Grid>
        </Grid>
    )
}

// see https://material-ui.com/components/buttons/#third-party-routing-library
const AdapterLink = React.forwardRef<HTMLAnchorElement, RouterLinkProps>((props, ref) => (
    <HashLink
        innerRef={ref}
        {...props}
        scroll={(e) => {
            // don't use the default scroll function, as scrollIntoView will hide the element under the appbar
            const y = e.getBoundingClientRect().top + window.pageYOffset - 70
            window.scrollTo({ top: y, behavior: 'smooth' })
        }}
    />
))
AdapterLink.displayName = 'AdapterLink'

interface ListItemLinkProps {
    children: JSX.Element | JSX.Element[]
    to: string
    listItemProps?: any // tried and failed - this should be the type of props in ListItem
    className?: string
}

// See https://material-ui.com/guides/composition/#routing-libraries
export function ListItemButtonLink(props: ListItemLinkProps) {
    const routerLink = React.forwardRef<HTMLAnchorElement, Omit<RouterLinkProps, 'innerRef' | 'to'>>(
        (itemProps, ref) => (
            // With react-router-dom@^6.0.0 use `ref` instead of `innerRef`
            // See https://github.com/ReactTraining/react-router/issues/6056
            <RouterLink to={props.to} {...itemProps} innerRef={ref} />
        )
    )
    routerLink.displayName = 'RouterLink'
    const renderLink = React.useMemo(() => routerLink, [routerLink])

    return (
        <ListItem {...props.listItemProps} button component={renderLink}>
            {props.children}
        </ListItem>
    )
}

// see https://material-ui.com/components/buttons/#third-party-routing-library
export function ListItemLink(props: ListItemLinkProps) {
    return <ListItem button component={AdapterLink} {...props} />
}

function hasMinHeight(o: any): o is CSSProperties {
    return o && (o as CSSProperties).minHeight !== undefined
}

// Returns a function that given a theme returns styles based on the toolbar min-height
// See https://stackoverflow.com/questions/45396236/material-ui-appbar-strategy-for-restricting-an-image-height-to-appbar-height
export function toolbarRelativeProperties(property: string, modifier = (value: any) => value) {
    return (theme: Theme) =>
        Object.keys(theme.mixins.toolbar).reduce((style, key) => {
            const value = theme.mixins.toolbar[key]
            if (key === 'minHeight') {
                return { ...style, [property]: modifier(value) }
            }
            if (hasMinHeight(value)) {
                return { ...style, [key]: { [property]: modifier(value.minHeight) } }
            }
            return style
        }, {})
}
