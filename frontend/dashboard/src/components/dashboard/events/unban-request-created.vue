<script lang="ts" setup>
import { IconUserQuestion } from '@tabler/icons-vue'

import Base from './base.vue'
import UserLink from './user-link.vue'

defineProps<{
	userName?: string | null
	userLogin?: string | null
	message?: string | null
	createdAt: string
}>()
</script>

<template>
	<Base
		v-if="userLogin && userName && message"
		:icon="IconUserQuestion"
		:icon-color="['#ff4f4d', '#ffaaa8']"
		:created-at="createdAt"
	>
		<template #leftContent>
			<div class="flex flex-col">
				<span>
					<UserLink :name="userLogin" :display-name="userName" />{{ '' }}
					<span class="font-bold">requested unban</span>
				</span>
				<span class="text-xs">{{ message }}</span>
			</div>
		</template>
	</Base>
</template>
