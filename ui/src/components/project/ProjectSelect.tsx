import React from 'react'
import Autocomplete from '@material-ui/lab/Autocomplete'
import { TextField } from '@material-ui/core'
import { useDispatch } from 'react-redux'
import { isEqual } from 'lodash-es'

import { useSelector } from '../../redux'
import { setDetail } from '../../redux/plans'
import { selectProjectNames } from '../../redux/projects'

interface OptionType {
    label: string
    value: string
}

export default function ProjectSelect(props: { planId: string; projectIds: string[] }) {
    const dispatch = useDispatch()
    const projectEntries: OptionType[] = useSelector(selectProjectNames, isEqual).map(({ id, name }) => ({
        label: name,
        value: id,
    }))
    const currentEntries = projectEntries.filter((v) => props.projectIds.includes(v.value))

    const handleChange = (_event: any, entries: OptionType[]) => {
        const projects = entries.map((v) => v.value)
        dispatch(setDetail({ id: props.planId, revisionDetails: { projects } }))
    }

    return (
        <Autocomplete
            multiple
            options={projectEntries}
            getOptionLabel={(option) => option.label}
            filterSelectedOptions
            renderInput={(params) => (
                <TextField
                    {...params}
                    required
                    variant="outlined"
                    label="Projects"
                    error={props.projectIds.length === 0}
                />
            )}
            value={currentEntries}
            onChange={handleChange}
            disableClearable
        />
    )
}
