import React, { useState } from 'react'
import { Chat } from '@berty-tech/hooks'
import './App.css'
import storage from 'redux-persist/lib/storage'

const CreateAccount: React.FC = () => {
	const [name, setName] = useState('')
	const [port, setPort] = useState(1337)
	const createAccount = Chat.useAccountCreate()
	return (
		<>
			<input
				type='text'
				placeholder='Account name'
				value={name}
				onChange={(e) => setName(e.target.value)}
			/>
			<input
				type='number'
				placeholder='Bridge port'
				value={port}
				onChange={(e) => setPort(parseInt(e.target.value, 10))}
			/>
			<button onClick={() => createAccount({ name, bridgePort: port })}>Create account</button>
		</>
	)
}

const Account: React.FC = () => {
	const account = Chat.useAccount()
	const link = Chat.useContactRequestReference()
	return (
		<>
			<h2>Account</h2>
			{account.name}
			<br />
			berty://{encodeURIComponent(link)}
		</>
	)
}

const AccountGate: React.FC = ({ children }) => {
	const account = Chat.useAccount()
	return account ? <>{children}</> : <CreateAccount />
}

const AddContact: React.FC = () => {
	const [link, setLink] = useState('')
	const sendContactRequest = Chat.useAccountSendContactRequest()
	const [, uriComponent] = link.split('berty://')
	const [b64Name, rdvSeed, pubKey] = uriComponent ? decodeURIComponent(uriComponent).split(' ') : []
	const name = b64Name && atob(b64Name)
	return (
		<>
			<h2>Add contact</h2>
			<input
				type='text'
				placeholder='Deep link'
				value={link}
				onChange={(e) => setLink(e.target.value)}
			/>
			{name}
			{!!(name && rdvSeed && pubKey) && (
				<button onClick={() => sendContactRequest(name, rdvSeed, pubKey)}>Add contact</button>
			)}
		</>
	)
}

const Contacts: React.FC = () => {
	const contacts = Chat.useAccountContacts()
	return (
		<>
			<h2>Contacts</h2>
			<div style={{ display: 'flex' }}>
				{contacts.map((contact) => (
					<div style={{ border: '1px solid black' }} key={contact.id}>
						{JSON.stringify(contact, null, 2)}
					</div>
				))}
			</div>
		</>
	)
}

const Message: React.FC<{ msgId: string }> = ({ msgId }) => {
	const message = Chat.useGetMessage(msgId)
	return <div style={{ textAlign: message.isMe ? 'left' : 'right' }}>{message.body}</div>
}

const Conversation: React.FC<{ convId: string }> = ({ convId }) => {
	const conv = Chat.useGetConversation(convId)
	const sendMessage = Chat.useMessageSend()
	const [message, setMessage] = useState('')
	return (
		!!conv && (
			<>
				{conv.title}
				<br />
				<div style={{ display: 'flex', flexDirection: 'column' }}>
					{conv.messages.map((msg) => (
						<Message key={msg} msgId={msg} />
					))}
				</div>
				<input
					type='text'
					placeholder='Type your message'
					value={message}
					onChange={(e) => setMessage(e.target.value)}
				/>
				{!!message && (
					<button onClick={() => sendMessage({ type: 'UserMessage', body: message, id: convId })}>
						Send
					</button>
				)}
			</>
		)
	)
}

const Conversations: React.FC = () => {
	const [selected, setSelected] = useState('')
	const conversations = Chat.useConversationList()
	return (
		<>
			<h2>Conversations</h2>
			<div style={{ display: 'flex' }}>
				{conversations.map((conv) => {
					if (conv.kind === 'fake') return null
					const isSelected = selected === conv.id
					return (
						<div
							style={{ border: '1px solid black', backgroundColor: isSelected && 'lightgrey' }}
							key={conv.id}
						>
							{conv.title}
							<br />
							<button onClick={() => setSelected(conv.id)}>Select</button>
						</div>
					)
				})}
			</div>
			{!!selected && <Conversation convId={selected} />}
		</>
	)
}

const ClearStorage: React.FC = () => (
	<button
		onClick={() => {
			localStorage.clear()
			window.location.reload()
		}}
	>
		Clear storage
	</button>
)

function App() {
	return (
		<Chat.Provider config={{ storage }}>
			<div className='App' style={{ display: 'flex', flexDirection: 'column' }}>
				<h1>Berty web dev</h1>
				<ClearStorage />
				<AccountGate>
					<Account />
					<AddContact />
					<Contacts />
					<Conversations />
				</AccountGate>
			</div>
		</Chat.Provider>
	)
}

export default App
