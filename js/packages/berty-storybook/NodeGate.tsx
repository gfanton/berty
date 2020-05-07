import React from 'react'
import { Chat } from '@berty-tech/hooks'
import { Text, View, ActivityIndicator } from 'react-native'

const NodeGate: React.FC = ({ children }) => {
	const client = Chat.useClient()
	const account = Chat.useAccount()
	return account?.onboarded && !client ? (
		<View style={{ height: '100%', justifyContent: 'center', alignItems: 'center' }}>
			<Text style={{ marginBottom: 20 }}>Starting Berty node..</Text>
			<ActivityIndicator size='large' />
		</View>
	) : (
		<>{children}</>
	)
}

export default NodeGate
