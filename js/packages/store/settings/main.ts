import { createSlice } from '@reduxjs/toolkit'
import { composeReducers } from 'redux-compose'
import { all, select } from 'redux-saga/effects'
import { makeDefaultReducers, makeDefaultCommandsSagas } from '../utils'
import { BertyNodeConfig } from '../protocol/client'

export type Entity = {
	id: string
	nodeConfig: BertyNodeConfig
}

export type State = { [key: string]: Entity }

export type GlobalState = {
	settings: {
		main: State
	}
}

export type Commands = {
	create: (state: State, action: { payload: Entity }) => State
	set: (state: State, action: { payload: { id: string; key: string; value: any } }) => State
	delete: (state: State, action: { payload: { id: string } }) => State
}

export type Queries = {
	get: (state: GlobalState, payload: { id: string }) => Entity | undefined
	getAll: (state: GlobalState) => Entity[]
}

export type Events = {
	created: (state: State, action: { payload: Entity }) => State
}

export function* getMainSettings(id: string) {
	const mainSettings = (yield select((state: GlobalState) => queries.get(state, { id }))) as
		| Entity
		| undefined
	if (!mainSettings) {
		throw new Error(`main settings not found for ${id}`)
	}
	return mainSettings
}

export type Transactions = {
	[K in keyof Commands]: Commands[K] extends (
		state: State,
		action: { payload: infer TPayload },
	) => State
		? (payload: TPayload) => Generator
		: never
} & {
	// put custom transactions here
}

const initialState: State = {}

const commandsNames = ['create', 'set', 'delete']

const commandHandler = createSlice<State, Commands>({
	name: 'protocol/client/command',
	initialState,
	// we don't change state on commands
	reducers: makeDefaultReducers(commandsNames),
})

const eventsNames = ['created'] as string[]

const eventHandler = createSlice<State, Events>({
	name: 'protocol/client/event',
	initialState,
	reducers: {
		...makeDefaultReducers(eventsNames),
		created: (state, { payload }) => {
			const { id } = payload
			if (!state[id]) {
				state[id] = payload
			}
			return state
		},
		// put reducer implementaion here
	},
})

export const reducer = composeReducers(commandHandler.reducer, eventHandler.reducer)
export const commands = commandHandler.actions
export const events = eventHandler.actions
export const queries: Queries = {
	get: (state, { id }) => state.settings.main[id],
	getAll: (state) => Object.values(state.settings.main),
}

export const transactions: Transactions = {
	create: function*() {},
	set: function*() {},
	delete: function*() {},
}

export function* orchestrator() {
	yield all([...makeDefaultCommandsSagas(commands, transactions)])
}
