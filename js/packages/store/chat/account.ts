import { createSlice, CaseReducer, PayloadAction } from '@reduxjs/toolkit'
import { composeReducers } from 'redux-compose'
import { fork, put, all, select, takeEvery, take } from 'redux-saga/effects'
import faker from 'faker'
import { Buffer } from 'buffer'
import { simpleflake } from 'simpleflakes/lib/simpleflakes-legacy'
import { berty } from '@berty-tech/api'
import { makeDefaultReducers, makeDefaultCommandsSagas, strToBuf, jsonToBuf } from '../utils'

import { contact, conversation } from '../chat'
import * as protocol from '../protocol'

export type Entity = {
	id: string
	name: string
	requests: Array<string>
	conversations: Array<string>
	contacts: Array<string>
	onboarded: boolean
	bridgePort: number
}

export type Event = {
	id: string
	version: number
	aggregateId: string
}

export type State = {
	aggregates: { [key: string]: Entity }
}

export type GlobalState = {
	chat: {
		account: State
	}
}

export namespace Command {
	export type Generate = void
	export type Create = { name: string; bridgePort: number }
	export type Delete = { id: string }
	export type SendContactRequest = {
		id: string
		contactName: string
		contactRdvSeed: string
		contactPublicKey: string
	}
	export type Replay = { id: string }
	export type Open = { id: string; bridgePort: number }
	export type Onboard = { id: string }
}

export namespace Query {
	export type List = {}
	export type Get = { id: string }
	export type GetRequestReference = { id: string }
	export type GetAll = void
	export type GetLength = undefined
}

export namespace Event {
	export type Created = {
		aggregateId: string
		name: string
		bridgePort: number
	}
	export type Deleted = { aggregateId: string }
	export type Onboarded = { aggregateId: string }
}

type SimpleCaseReducer<P> = CaseReducer<State, PayloadAction<P>>

export type CommandsReducer = {
	generate: SimpleCaseReducer<Command.Generate>
	create: SimpleCaseReducer<Command.Create>
	delete: SimpleCaseReducer<Command.Delete>
	sendContactRequest: SimpleCaseReducer<Command.SendContactRequest>
	replay: SimpleCaseReducer<Command.Replay>
	open: SimpleCaseReducer<Command.Open>
	onboard: SimpleCaseReducer<Command.Onboard>
}

export type QueryReducer = {
	list: (state: GlobalState, query: Query.List) => Array<Entity>
	get: (state: GlobalState, query: Query.Get) => Entity
	getRequestRdvSeed: (
		state: protocol.client.GlobalState,
		query: Query.GetRequestReference,
	) => string | undefined
	getAll: (state: GlobalState, query: Query.GetAll) => Array<Entity>
	getLength: (state: GlobalState) => number
}

export type EventsReducer = {
	created: SimpleCaseReducer<Event.Created>
	deleted: SimpleCaseReducer<Event.Deleted>
	onboarded: SimpleCaseReducer<Event.Onboarded>
}

export type Transactions = {
	[K in keyof CommandsReducer]: CommandsReducer[K] extends SimpleCaseReducer<infer TPayload>
		? (payload: TPayload) => Generator
		: never
}

const initialState = {
	aggregates: {},
}

const commandHandler = createSlice<State, CommandsReducer>({
	name: 'chat/account/command',
	initialState,
	// this is stupid but getting https://github.com/kimamula/ts-transformer-keys in might be a headache
	// maybe move commands and events definitions in .protos
	reducers: makeDefaultReducers([
		'generate',
		'create',
		'delete',
		'sendContactRequest',
		'replay',
		'open',
		'onboard',
	]),
})

const eventHandler = createSlice<State, EventsReducer>({
	name: 'chat/account/event',
	initialState,
	reducers: {
		created: (state, { payload }) => {
			if (!state.aggregates[payload.aggregateId]) {
				state.aggregates[payload.aggregateId] = {
					id: payload.aggregateId,
					name: payload.name,
					requests: [],
					conversations: [],
					contacts: [],
					onboarded: false,
					bridgePort: payload.bridgePort,
				}
			}
			return state
		},
		deleted: (state, { payload }) => {
			delete state.aggregates[payload.aggregateId]
			return state
		},
		onboarded: (state, { payload }) => {
			const account = state.aggregates[payload.aggregateId]
			if (account) {
				account.onboarded = true
			}
			return state
		},
	},
})

export const reducer = composeReducers(commandHandler.reducer, eventHandler.reducer)
export const commands = commandHandler.actions
export const events = eventHandler.actions
export const queries: QueryReducer = {
	list: (state) => Object.values(state.chat.account.aggregates),
	get: (state, { id }) => state.chat.account.aggregates[id],
	getRequestRdvSeed: (state, { id }) =>
		protocol.queries.client.get(state, { id })?.contactRequestRdvSeed,
	getAll: (state) => Object.values(state.chat.account.aggregates),
	getLength: (state) => Object.keys(state.chat.account.aggregates).length,
}

export const getProtocolClient = function*(id: string): Generator<unknown, protocol.Client, void> {
	const client = (yield select((state) => protocol.queries.client.get(state, { id }))) as
		| protocol.Client
		| undefined
	if (client == null) {
		throw new Error('client is not defined')
	}
	return client
}

export const transactions: Transactions = {
	open: function*({ id, bridgePort }) {
		yield* protocol.transactions.client.start({ id, bridgePort })
		yield* conversation.transactions.open({ accountId: id })

		const client = yield* getProtocolClient(id)

		yield* protocol.transactions.client.listenToGroupMetadata({
			clientId: id,
			groupPk: strToBuf(client.accountGroupPk),
		})

		yield* protocol.transactions.client.contactRequestReference({ id })
	},
	generate: function*() {
		yield* transactions.create({ name: faker.name.firstName(), bridgePort: 1337 })
	},
	create: function*({ name, bridgePort }) {
		// create an id for the account
		const id = simpleflake().toString()

		const event = events.created({
			aggregateId: id,
			name,
			bridgePort,
		})
		// open account
		yield* transactions.open({ id, bridgePort })
		// get account PK

		yield* protocol.transactions.client.contactRequestResetReference({ id })
		yield* protocol.transactions.client.contactRequestEnable({ id })

		yield put(event)

		/*const client = yield* getProtocolClient(id)

		yield* protocol.transactions.client.appMetadataSend({
			id,
			groupPk: strToBuf(client.accountGroupPk),
			payload: jsonToBuf(event),
		})*/
	},
	delete: function*() {
		yield put({ type: 'CLEAR_STORE' })
	},
	replay: function*({ id }) {
		const account = select((state) => queries.get(state, { id }))
		if (account == null) {
			console.error('account does not exist')
			return
		}

		const client = yield* getProtocolClient(id)

		// replay log from first event
		const chan = yield* protocol.transactions.client.groupMetadataSubscribe({
			id: client.id,
			groupPk: new Buffer(client.accountGroupPk),
			// TODO: use last cursor
			since: new Uint8Array(),
			until: new Uint8Array(),
			goBackwards: false,
		})
		yield takeEvery(chan, function*(
			action: protocol.client.Commands extends {
				[key: string]: (
					state: protocol.client.State,
					action: infer UAction,
				) => protocol.client.State
			}
				? UAction
				: { type: string; payload: { event: any } },
		) {
			yield put(action)
			if (action.type === protocol.events.client.groupMetadataPayloadSent.type) {
				yield put(action.payload.event)
			}
		})
	},
	sendContactRequest: function*(payload) {
		const account = (yield select((state) => queries.get(state, { id: payload.id }))) as
			| Entity
			| undefined
		if (account == null) {
			throw new Error("account doesn't exist")
		}

		const metadata = {
			name: account.name,
			givenName: payload.contactName,
		}

		const contact: berty.types.IShareableContact = {
			pk: strToBuf(payload.contactPublicKey),
			publicRendezvousSeed: strToBuf(payload.contactRdvSeed),
			metadata: jsonToBuf(metadata),
		}

		yield* protocol.transactions.client.contactRequestSend({
			id: payload.id,
			contact,
		})
	},
	onboard: function*(payload) {
		yield put(events.onboarded({ aggregateId: payload.id }))
	},
}

export function* orchestrator() {
	yield all([
		...makeDefaultCommandsSagas(commands, transactions),
		fork(function*() {
			yield take('persist/REHYDRATE')
			// start protocol clients
			const accounts = (yield select(queries.getAll)) as Entity[]
			for (const account of accounts) {
				yield* transactions.open({ id: account.id, bridgePort: account.bridgePort })
				yield* contact.transactions.open({ accountId: account.id })
			}
		}),
	])
}
