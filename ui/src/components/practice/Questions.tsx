import React, { useState, useEffect } from 'react'
import {
    Radio,
    RadioGroup,
    FormControlLabel,
    TextField,
    List,
    ListItem,
    ListItemText,
    makeStyles,
    Theme,
    createStyles,
} from '@material-ui/core'
import { ArrowRight } from '@material-ui/icons'
import Markdown from '../common/MaterialMarkdown'

import * as api from '../../client'
import { useDispatch } from 'react-redux'
import { answerQuestion, selectIsPlanReadonly, selectPlanRevision } from '../../redux/plans'
import { useSelector } from '../../redux'

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        listItemIcon: { marginBottom: '4px' },
    })
)

export function QuestionList(props: { qs: api.IQuestion[]; practiceId: string; planId?: string; revisionId?: string }) {
    const classes = useStyles()
    return (
        <List>
            {props.qs.map((q, i) => {
                if (props.planId) {
                    return (
                        <PlanQuestion
                            key={q.id}
                            q={q}
                            practiceId={props.practiceId}
                            qId={q.id!}
                            planId={props.planId}
                            revisionId={props.revisionId}
                        />
                    )
                } else {
                    return (
                        <ListItem key={q.id} dense>
                            <ArrowRight className={classes.listItemIcon} />
                            <ListItemText>
                                <Markdown paraVariant="body1" source={q.text} />
                            </ListItemText>
                        </ListItem>
                    )
                }
            })}
        </List>
    )
}

function PlanQuestion(props: {
    q: api.IQuestion
    practiceId: string
    qId: string
    planId: string
    revisionId?: string
}) {
    const a = useSelector(
        (state) =>
            selectPlanRevision(state, props.planId, props.revisionId).answersById?.[props.practiceId]?.[props.qId] ?? {
                answer: api.Answer2.Unanswered,
            }
    )
    const readOnly = useSelector((state) => selectIsPlanReadonly(state, props.planId, props.revisionId))
    const dispatch = useDispatch()
    const [notesValue, setNotes] = useState(a.notes) // we keep the note state local until the user moves away
    useEffect(() => {
        setNotes(a.notes)
    }, [a.notes])

    const save = (answer: Partial<api.IAnswer>) => {
        if (!readOnly) {
            dispatch(
                answerQuestion({
                    id: props.planId,
                    pId: props.practiceId,
                    qId: props.qId,
                    value: Object.assign({}, a, answer),
                })
            )
        }
    }

    let na = props.q.na
    if (na === undefined) {
        na = true
    }
    return (
        <ListItem alignItems="flex-start">
            <ListItemText>
                <Markdown source={props.q.text} />
                <YesNoNA na={na} value={a.answer} checked={(answer) => save({ answer })} />
                <TextField
                    id={'notes-' + props.qId}
                    label="Notes"
                    value={notesValue}
                    onChange={(event) => setNotes(event.target.value)}
                    onBlur={() => save({ notes: notesValue })}
                    InputProps={{ readOnly: readOnly }}
                    variant="outlined"
                    multiline
                    fullWidth
                    margin="normal"
                />
            </ListItemText>
        </ListItem>
    )
}

const YesNoValues = [api.Answer2.Yes, api.Answer2.No]
const YesNoNAValues = YesNoValues.concat([api.Answer2.N_A])
function YesNoNA(props: { na: boolean; value: api.Answer2; checked: (answer: api.Answer2) => void }) {
    const values = props.na ? YesNoNAValues : YesNoValues

    return (
        <RadioGroup aria-label="position" name="position" row>
            {values.map((v) => (
                <FormControlLabel
                    key={v}
                    value={v}
                    checked={props.value === v}
                    onChange={(_event, checked) => {
                        if (checked) {
                            props.checked(v)
                        }
                    }}
                    control={<Radio color="primary" />}
                    label={v}
                    labelPlacement="end"
                />
            ))}
        </RadioGroup>
    )
}
