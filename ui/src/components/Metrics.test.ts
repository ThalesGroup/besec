import { PracticeMaturity, TaskResult } from './Metrics'
import { ExtendedPractice } from '../redux/practices'
import { Task, IAnswer, Answer2 } from '../client'

const fourtasks: ExtendedPractice = {
    id: 'a',
    name: 'a',
    allQuestionIds: ['q1', 'q2', 'q3', 'q4'],
    tasks: [
        { id: 't1', level: 1, questions: [{ id: 'q1' }] },
        { id: 't2', level: 2, questions: [{ id: 'q2a' }, { id: 'q2b' }] },
        { id: 't3', level: 3, questions: [{ id: 'q3' }] },
        { id: 't4', level: 4, questions: [{ id: 'q4' }] }
    ] as Task[]
}
const onetask: ExtendedPractice = { ...fourtasks, tasks: fourtasks.tasks.slice(3) }

const fouryeses: Record<string, IAnswer> = {
    q1: { answer: Answer2.Yes },
    q2a: { answer: Answer2.Yes },
    q2b: { answer: Answer2.Yes },
    q3: { answer: Answer2.Yes },
    q4: { answer: Answer2.Yes }
}
const yesAndNA: Record<string, IAnswer> = {
    q1: { answer: Answer2.N_A },
    q2a: { answer: Answer2.Yes },
    q2b: { answer: Answer2.N_A },
    q3: { answer: Answer2.Yes },
    q4: { answer: Answer2.N_A }
}
const oneyes: Record<string, IAnswer> = { q4: fouryeses['q4'] }
const nol1: Record<string, IAnswer> = {
    q1: { answer: Answer2.No },
    q2a: { answer: Answer2.Yes },
    q2b: { answer: Answer2.N_A },
    q3: { answer: Answer2.Yes },
    q4: { answer: Answer2.Yes }
}
const unans: Record<string, IAnswer> = { ...fouryeses, q3: { answer: Answer2.Unanswered } }

describe('PracticeMaturity', () => {
    it('scores 4 if all tasks are done', () => {
        expect(PracticeMaturity(fourtasks, fouryeses)).toBe(4)
    })

    it('scores 4 if all tasks are done or NA', () => {
        expect(PracticeMaturity(fourtasks, yesAndNA)).toBe(4)
    })

    it('scores 4 if the only task is done', () => {
        expect(PracticeMaturity(onetask, oneyes)).toBe(4)
    })

    it("scores 0 if level 1 isn't done", () => {
        expect(PracticeMaturity(fourtasks, nol1)).toBe(0)
    })

    it('returns -1 if a question is unanswered', () => {
        expect(PracticeMaturity(fourtasks, unans)).toBe(-1)
    })
})

describe('TaskResult', () => {
    it.each([Answer2.Yes, Answer2.No, Answer2.N_A, Answer2.Unanswered])(
        'acts as the identity function for single answer questions',
        ans => {
            expect(TaskResult(fourtasks.tasks[0], { q1: { answer: ans } } as Record<string, IAnswer>)).toBe(ans)
        }
    )
    it('treats a mix of N/A and Yes as Yes', () => {
        expect(
            TaskResult(fourtasks.tasks[1], {
                q2a: { answer: Answer2.Yes },
                q2b: { answer: Answer2.N_A }
            } as Record<string, IAnswer>)
        ).toBe(Answer2.Yes)
    })
    it('treats a mix of No and Yes as No', () => {
        expect(
            TaskResult(fourtasks.tasks[1], {
                q2a: { answer: Answer2.No },
                q2b: { answer: Answer2.Yes }
            } as Record<string, IAnswer>)
        ).toBe(Answer2.No)
    })
    it('treats a mix of Yes and Unanswered as Unanswered', () => {
        expect(
            TaskResult(fourtasks.tasks[1], {
                q2a: { answer: Answer2.Yes },
                q2b: { answer: Answer2.Unanswered }
            } as Record<string, IAnswer>)
        ).toBe(Answer2.Unanswered)
    })
})
