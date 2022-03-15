import React from 'react'
import {
    Dialog,
    Tooltip,
    Button,
    CircularProgress,
    createStyles,
    makeStyles,
    Theme,
    DialogContent,
    DialogContentText,
    DialogActions
} from '@material-ui/core'

import { ButtonProps } from '@material-ui/core/Button'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        button: {
            margin: theme.spacing(1)
        },
        rightIcon: {
            marginLeft: theme.spacing(1),
            height: '1em'
        },
        icon: {
            fontSize: 20
        },
        iconVariant: {
            opacity: 0.9,
            marginRight: theme.spacing(1)
        },
        message: {
            display: 'flex',
            alignItems: 'center'
        },
        error: {
            backgroundColor: theme.palette.error.dark
        }
    })
)

function OptionalTooltip(props: { children: any; title?: string }) {
    if (props.title) {
        return (
            <Tooltip title={props.title} placement="top-start">
                {props.children}
            </Tooltip>
        )
    } else {
        return props.children
    }
}

export interface APIButtonProps {
    tooltipTitle?: string // if not provided or is an empty string, no tooltip is rendered. Otherwise a tooltip is rendered even if the button is disabled.
    btnText: string
    img?: any
    className?: ButtonProps['className']
    buttonProps?: Partial<ButtonProps>
    // if present, a confirmation dialogue will be presented to the user first
    confirmation?: { verb: string; question: string }
    onClick: () => void
    inProgress: boolean
}
export default function APIButton(props: APIButtonProps) {
    const classes = useStyles()
    const [dialogOpen, setDialogOpen] = React.useState(false)

    const dialogCancel = () => {
        setDialogOpen(false)
    }

    let onButtonClick = props.onClick
    if (props.confirmation) {
        // If we need confirmation, the action is triggered from the dialog's button, instead of the original button
        onButtonClick = () => setDialogOpen(true)
    }

    return (
        <>
            <OptionalTooltip title={props.tooltipTitle}>
                <span className={`${classes.button} ${props.className}`}>
                    <Button {...props.buttonProps} onClick={onButtonClick}>
                        {props.btnText}
                        {!props.inProgress && props.img}
                        {props.inProgress && <CircularProgress size="1em" className={classes.rightIcon} />}
                    </Button>
                </span>
            </OptionalTooltip>
            {props.confirmation && (
                <Dialog
                    open={dialogOpen}
                    onClose={dialogCancel}
                    aria-labelledby="alert-dialog-title"
                    aria-describedby="alert-dialog-description"
                >
                    <DialogContent>
                        <DialogContentText id="alert-dialog-description">
                            {props.confirmation.question}
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={dialogCancel} color="primary">
                            Cancel
                        </Button>
                        <Button onClick={props.onClick} color="primary" autoFocus>
                            {props.confirmation.verb}
                        </Button>
                    </DialogActions>
                </Dialog>
            )}
        </>
    )
}
APIButton.defaultProps = { buttonProps: { color: 'default', size: 'medium', variant: 'text' } }
