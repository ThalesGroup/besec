// Originally from https://material-ui.com/components/drawers/
// Also see https://mui-treasury.com/layout-builder

import React, { useEffect } from 'react'
import {
    Drawer,
    List,
    ListItemIcon,
    ListItemText,
    makeStyles,
    Hidden,
    Theme,
    createStyles,
    AppBar,
    Toolbar,
    IconButton,
    Typography,
    CssBaseline,
    Paper,
    Grid,
} from '@material-ui/core'
import { Home, ShowChart, Ballot, Book, Menu } from '@material-ui/icons'
import { useLocation } from 'react-router-dom'

import { ListItemLink, toolbarRelativeProperties } from './common/common'
import { useSelector } from '../redux'
import { selectLoggedIn } from '../redux/session'
import { UserMenu } from './Auth'

// to work around https://github.com/facebook/create-react-app/issues/11770
// minimizing didn't work, so using this webpack loader directive instead, despite all the eslint warnings
import logo from '!file-loader!../images/logo.svg' // eslint-disable-line

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        logo: {
            width: '80px',
            marginRight: theme.spacing(2),
        },
        menuButton: {
            marginRight: theme.spacing(2),
            [theme.breakpoints.up('md')]: {
                display: 'none',
            },
        },
        drawerPaper: {
            // Setting height (e.g. with 'toolbarRelativeProperties('height', (h)=> `calc(100% - $h)`) doesn't work well, because every parent needs to have an explicit height set (e.g. to 100%)
            // instead we use position: sticky. Sticky means it will fill the grid column's width, whereas fixed won't.
            position: 'sticky',
            ...toolbarRelativeProperties('top')(theme),
        },
        main: {
            ...toolbarRelativeProperties('marginTop')(theme),
            padding: theme.spacing(2),
            minHeight: '300px', // without this, really tiny pages mean the body is too small to allow the sticky drawer to position itself below the top of the page
        },
    })
)

export default function Navigation(props: { children?: any }) {
    const classes = useStyles()
    const [mobileOpen, setMobileOpen] = React.useState(false)

    const loggedIn = useSelector(selectLoggedIn)

    const menuItems: [string, string, JSX.Element][] = [
        ['About', '/', <Home key="about" />],
        ['Plans', '/plans', <Ballot key="plans" />],
        ['Practices', '/practices', <Book key="practices" />],
        ['Metrics', '/metrics', <ShowChart key="metrics" />],
    ]

    const drawerContents = (
        <nav>
            <List>
                {menuItems.map(([text, path, icon]) => (
                    <ListItemLink key={text} to={path}>
                        <ListItemIcon>{icon}</ListItemIcon>
                        <ListItemText>{text}</ListItemText>
                    </ListItemLink>
                ))}
            </List>
        </nav>
    )

    function handleDrawerToggle() {
        setMobileOpen(!mobileOpen)
    }

    const appBar = (
        <AppBar position="fixed">
            <Toolbar>
                <IconButton
                    color="inherit"
                    aria-label="open drawer"
                    edge="start"
                    onClick={handleDrawerToggle}
                    className={classes.menuButton}
                >
                    <Menu />
                </IconButton>

                <div className={classes.logo}>
                    <a href={process.env.PUBLIC_URL}>
                        <img src={logo} alt="Logo" className={classes.logo} />
                    </a>
                </div>
                <Typography variant="h6" noWrap>
                    BeSec: {/*<tagline>*/}
                </Typography>
                {loggedIn && <UserMenu />}
            </Toolbar>
        </AppBar>
    )

    // We show a temporary drawer when small, for mobile users. Otherwise it's not a Drawer, but a side panel.
    // Tried to put the Drawer into a Grid-item, so we could use Grid to keep the other elements off the Drawer.
    // It doesn't work - I assume the Drawer CSS is not compatible. This meant that the main content had to know about
    // and try and avoid the drawer.
    const mobileNav = (
        <>
            <Drawer
                variant="temporary"
                open={mobileOpen}
                onClose={handleDrawerToggle}
                classes={{
                    paper: classes.drawerPaper,
                }}
                ModalProps={{
                    keepMounted: true, // Better open performance on mobile.
                }}
            >
                {drawerContents}
            </Drawer>
            <main className={classes.main}>{props.children}</main>
        </>
    )

    const largeNav = (
        <Grid container>
            <Grid item md={2} xl={1}>
                <Paper className={classes.drawerPaper}>{drawerContents}</Paper>
            </Grid>
            <Grid item md={10} xl={11}>
                <main className={classes.main}>{props.children}</main>
            </Grid>
        </Grid>
    )

    return (
        <>
            <CssBaseline />
            <ScrollToTop />

            {appBar}

            <Hidden mdUp>{mobileNav}</Hidden>
            <Hidden smDown>{largeNav}</Hidden>
        </>
    )
}

// See https://reacttraining.com/react-router/web/guides/scroll-restoration
function ScrollToTop() {
    const { pathname } = useLocation()

    useEffect(() => {
        window.scrollTo(0, 0)
    }, [pathname])

    return null
}
