import React, { useState } from 'react'
import { TouchableOpacity, View, Platform, TextInput, StyleSheet } from 'react-native'
import { Text, Icon } from '@ui-kitten/components'
import { BlurView } from '@react-native-community/blur'
import { CommonActions } from '@react-navigation/native'
import { useTranslation } from 'react-i18next'
import EmojiBoard from 'react-native-emoji-board'

import { KeyboardAvoidingView } from '@berty-tech/components/shared-components/KeyboardAvoidingView'
import { useStyles } from '@berty-tech/styles'
import { Routes, ScreenProps, useNavigation } from '@berty-tech/navigation'
import {
	useConversation,
	useLastConvInteraction,
	useReadEffect,
	useNotificationsInhibitor,
	useMsgrContext,
} from '@berty-tech/store/hooks'
import beapi from '@berty-tech/api'

import { ChatFooter, ChatDate } from './common'
import { SwipeNavRecognizer } from '../shared-components/SwipeNavRecognizer'
import { useLayout } from '../hooks'
import { MultiMemberAvatar } from '../avatars'
import { MessageList } from '@berty-tech/components/chat/MessageList'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ReplyReactionProvider } from './ReplyReactionContext'
import AndroidKeyboardAdjust from 'react-native-android-keyboard-adjust'
import { useFocusEffect } from '@react-navigation/native'

//
// MultiMember
//

// Styles

const HeaderMultiMember: React.FC<{
	id: string
	stickyDate?: number
	showStickyDate?: boolean
}> = ({ id, stickyDate, showStickyDate }) => {
	const [isEdit, setIsEdit] = useState(false)
	const ctx = useMsgrContext()
	const { navigate, goBack } = useNavigation()
	const insets = useSafeAreaInsets()
	const [{ row, padding, flex, text, color, border }, { scaleSize }] = useStyles()
	const conversation = useConversation(id)
	const [editValue, setEditValue] = useState(conversation?.displayName || '')
	const [layoutHeader, onLayoutHeader] = useLayout() // to position date under blur

	const editDisplayName = async () => {
		const buf = beapi.messenger.AppMessage.SetGroupInfo.encode({ displayName: editValue }).finish()
		await ctx.client?.interact({
			conversationPublicKey: conversation?.publicKey,
			type: beapi.messenger.AppMessage.Type.TypeSetGroupInfo,
			payload: buf,
		})
		setIsEdit(false)
	}
	return (
		<View style={{ position: 'absolute', top: 0, left: 0, right: 0 }} onLayout={onLayoutHeader}>
			<BlurView
				overlayColor=''
				blurType='light'
				blurAmount={30}
				style={{ position: 'absolute', bottom: 0, top: 0, left: 0, right: 0 }}
			/>
			<View
				style={[
					{
						alignItems: 'center',
						flexDirection: 'row',
						justifyContent: 'space-between',
						marginTop: insets.top,
					},
					padding.medium,
				]}
			>
				<TouchableOpacity style={[flex.tiny]} onPress={goBack}>
					<Icon
						name='arrow-back-outline'
						width={25 * scaleSize}
						height={25 * scaleSize}
						fill={color.black}
					/>
				</TouchableOpacity>
				{isEdit ? (
					<View
						style={[
							flex.medium,
							border.radius.small,
							{
								flexDirection: 'row',
								alignItems: 'center',
								backgroundColor: color.light.grey,
							},
						]}
					>
						<View style={[flex.medium]} />
						<TextInput
							style={[
								flex.large,
								text.color.black,
								text.align.center,
								text.bold.medium,
								text.size.scale(20),
								padding.vertical.small,
							]}
							autoFocus
							onSubmitEditing={editDisplayName}
							onBlur={() => {
								setIsEdit(false)
								setEditValue(conversation?.displayName || '')
							}}
							value={editValue}
							onChange={({ nativeEvent }) => setEditValue(nativeEvent.text)}
						/>
						<TouchableOpacity
							style={[flex.medium, { alignItems: 'flex-end' }]}
							onPress={editDisplayName}
						>
							<Icon name='checkmark-outline' height={25} width={25} fill={color.grey} />
						</TouchableOpacity>
					</View>
				) : (
					<TouchableOpacity style={[flex.large]} onLongPress={() => setIsEdit(true)}>
						<Text
							numberOfLines={1}
							style={[text.align.center, text.bold.medium, text.size.scale(20)]}
						>
							{conversation?.displayName || ''}
						</Text>
					</TouchableOpacity>
				)}

				<View style={[flex.tiny, row.fill, { alignItems: 'center' }]}>
					<TouchableOpacity
						style={[flex.small, row.right]}
						onPress={() => navigate.chat.groupSettings({ convId: id })}
					>
						<MultiMemberAvatar publicKey={conversation?.publicKey} size={40 * scaleSize} />
					</TouchableOpacity>
				</View>
			</View>
			{!!stickyDate && !!showStickyDate && layoutHeader?.height && (
				<View
					style={{
						position: 'absolute',
						top: layoutHeader.height + 10,
						left: 0,
						right: 0,
					}}
				>
					<ChatDate date={stickyDate} />
				</View>
			)}
		</View>
	)
}

const NT = beapi.messenger.StreamEvent.Notified.Type

export const MultiMember: React.FC<ScreenProps.Chat.Group> = ({ route: { params } }) => {
	useNotificationsInhibitor((_ctx, notif) => {
		if (
			notif.type === NT.TypeMessageReceived &&
			(notif.payload as any)?.payload?.interaction?.conversationPublicKey === params?.convId
		) {
			return 'sound-only'
		}
		return false
	})
	const [{ background, flex }] = useStyles()
	const { dispatch } = useNavigation()
	useReadEffect(params.convId, 1000)
	const conv = useConversation(params?.convId)
	const { t } = useTranslation()
	const ctx = useMsgrContext()
	const insets = useSafeAreaInsets()

	const lastInte = useLastConvInteraction(params?.convId || '')
	const lastUpdate = conv?.lastUpdate || lastInte?.sentDate || conv?.createdDate || null
	const [stickyDate, setStickyDate] = useState(lastUpdate || null)
	const [showStickyDate, setShowStickyDate] = useState(false)

	const [isSwipe, setSwipe] = useState(true)

	useFocusEffect(
		React.useCallback(() => {
			AndroidKeyboardAdjust?.setAdjustResize()
			return () => AndroidKeyboardAdjust?.setAdjustPan()
		}, []),
	)

	return (
		<ReplyReactionProvider>
			{({ activeEmojiKeyboardCid, setActiveEmojiKeyboardCid, setActivePopoverCid }) => {
				const onRemoveEmojiBoard = () => {
					setActivePopoverCid(activeEmojiKeyboardCid)
					setActiveEmojiKeyboardCid(null)
				}
				return (
					<View style={[flex.tiny, background.white]}>
						<SwipeNavRecognizer
							onSwipeLeft={() =>
								isSwipe &&
								dispatch(
									CommonActions.navigate({
										name: Routes.Chat.MultiMemberSettings,
										params: { convId: params?.convId },
									}),
								)
							}
						>
							<KeyboardAvoidingView
								style={[flex.tiny]}
								behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
								bottomFixedViewPadding={20}
							>
								<MessageList id={params?.convId} {...{ setStickyDate, setShowStickyDate }} />
								<ChatFooter
									convPk={params?.convId}
									placeholder={t('chat.multi-member.input-placeholder')}
									setSwipe={setSwipe}
								/>
								<HeaderMultiMember
									id={params?.convId}
									{...({ stickyDate, showStickyDate } as any)}
								/>
							</KeyboardAvoidingView>
						</SwipeNavRecognizer>
						{!!activeEmojiKeyboardCid && (
							<View style={StyleSheet.absoluteFill}>
								<TouchableOpacity
									style={[StyleSheet.absoluteFill, { flex: 1 }]}
									activeOpacity={0.9}
									onPress={onRemoveEmojiBoard}
								/>
								<EmojiBoard
									showBoard={true}
									onClick={(emoji) => {
										ctx.client
											?.interact({
												conversationPublicKey: conv?.publicKey,
												type: beapi.messenger.AppMessage.Type.TypeUserReaction,
												payload: beapi.messenger.AppMessage.UserReaction.encode({
													emoji: `:${emoji.name}:`,
													state: true,
												}).finish(),
												targetCid: activeEmojiKeyboardCid,
											})
											.then(() => {
												ctx.playSound('messageSent')
												setActivePopoverCid(null)
												setActiveEmojiKeyboardCid(null)
											})
											.catch((e) => {
												console.warn('e sending message:', e)
											})
									}}
									onRemove={() => {
										setActivePopoverCid(activeEmojiKeyboardCid)
										setActiveEmojiKeyboardCid(null)
									}}
									containerStyle={{
										position: 'absolute',
										bottom: 0,
										paddingBottom: insets.bottom,
									}}
								/>
							</View>
						)}
					</View>
				)
			}}
		</ReplyReactionProvider>
	)
}
