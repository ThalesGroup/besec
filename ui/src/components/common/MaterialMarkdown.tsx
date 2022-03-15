/*
 react-markdown uses <p> tags etc by default, we want to use material-ui components.

 Usage:
    import Markdown from './materialMarkdown'
    ...
    let input = "# Heading!"
    return <div>{Markdown(input})}</div>

 Adapted from https://github.com/data-driven-forms/react-forms/blob/master/packages/react-renderer-demo/src/editor-demo/common/md-helper/index.js
 Also see https://github.com/tadasant/strand-ui/blob/master/src/components/strand/common/markdownRenderers.tsx
*/

import React from 'react'
import ReactMarkdown from 'react-markdown'
import { Link, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@material-ui/core'
import { TypographyProps } from '@material-ui/core/Typography'

interface MarkdownProps {
    source?: string
    headingLevelOffset?: number
    paraVariant?: TypographyProps['variant']
    color?: TypographyProps['color']
}
export default function Markdown({
    source = '',
    headingLevelOffset = 2,
    paraVariant = 'body1',
    color = 'textPrimary',
}: MarkdownProps) {
    const heading = ({ level, children }: { level: number; children: any }) => (
        <Typography color={color} variant={`h${level + headingLevelOffset}` as TypographyProps['variant']}>
            {children}
        </Typography>
    )

    const components: { [tag: string]: React.ElementType<any> } = {
        p: ({ children }) => (
            <Typography variant={paraVariant} color={color} gutterBottom>
                {children}
            </Typography>
        ),
        a: ({ href, children }) => (
            <Link href={href} target="_blank" rel="noopener noreferrer">
                {children}
            </Link>
        ),
        h1: heading,
        h2: heading,
        h3: heading,
        h4: heading,
        h5: heading,
        h6: heading,
        // ol, ul, li // material-ui's ListItem isn't designed for bulleted lists
        table: ({ children }) => (
            <Paper style={{ marginBottom: 10, marginTop: 10 }}>
                <Table>{children}</Table>
            </Paper>
        ),
        tbody: ({ children }) => <TableBody>{children}</TableBody>,
        thead: ({ children }) => <TableHead>{children}</TableHead>,
        tr: ({ children }) => <TableRow>{children}</TableRow>,
        td: ({ children }) => <TableCell>{children}</TableCell>,
        th: ({ children }) => <TableCell>{children}</TableCell>,
    }

    return (
        <ReactMarkdown components={components} linkTarget="_blank">
            {source}
        </ReactMarkdown>
    )
}
