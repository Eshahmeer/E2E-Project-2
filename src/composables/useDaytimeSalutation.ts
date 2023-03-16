import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useNow} from '@vueuse/core'

import {useAuthStore} from '@/stores/auth'
import {hourToDaytime} from '@/helpers/hourToDaytime'

export type Daytime = 'night' | 'morning' | 'day' | 'evening'

export function useDaytimeSalutation() {
	const {t} = useI18n({useScope: 'global'})
	const now = useNow({ 
		interval: 30 * 60 * 1000, // Every 30 minutes
	})
	const authStore = useAuthStore()

	const name = computed(() => authStore.userDisplayName)
	const daytime = computed(() => hourToDaytime(now.value))

	const salutations = {
		'night': () => t('home.welcomeNight', {username: name.value}),
		'morning': () => t('home.welcomeMorning', {username: name.value}),
		'day': () => t('home.welcomeDay', {username: name.value}),
		'evening': () => t('home.welcomeEvening', {username: name.value}),
	} as Record<Daytime, () => string>

	return computed(() => name.value ? salutations[daytime.value]() : undefined)
}