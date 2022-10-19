import {computed, ref, shallowReactive, watch, type Ref} from 'vue'
import cloneDeep from 'lodash.clonedeep'

import type {Filter} from '@/composables/useRouteFilter'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import TaskCollectionService, {type GetAllTasksParams} from '@/services/taskCollection'
import TaskService from '@/services/task'

import TaskModel from '@/models/task'
import {error, success} from '@/message'

// FIXME: unify with general `useTaskList`
export function useGanttTaskList<F extends Filter>(
	filters: Ref<F>,
	filterToApiParams: (filters: F) => GetAllTasksParams,
	options: {
		loadAll?: boolean,
	} = {
		loadAll: true,
	}) {
	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const taskService = shallowReactive(new TaskService())

	const isLoading = computed(() => taskCollectionService.loading)

	const tasks = ref<Map<ITask['id'], ITask>>(new Map())

	async function fetchTasks(params: GetAllTasksParams, page = 1): Promise<ITask[]> {
		const tasks = await taskCollectionService.getAll({listId: filters.value.listId}, params, page) as ITask[]
		if (options.loadAll && page < taskCollectionService.totalPages) {
			const nextTasks = await fetchTasks(params, page + 1)
			return tasks.concat(nextTasks)
		}
		return tasks
	}

	/**
	 * Load and assign new tasks
	 * Normally there is no need to trigger this manually
	 */
	async function loadTasks() {
		const params: GetAllTasksParams = filterToApiParams(filters.value)

		const loadedTasks = await fetchTasks(params)
		tasks.value = new Map()
		loadedTasks.forEach(t => tasks.value.set(t.id, t))
	}

	/**
	 * Load tasks when filters change
	 */
	watch(
		filters,
		() => loadTasks(),
		{immediate: true, deep: true},
	)

	async function addTask(task: Partial<ITask>) {
		const newTask = await taskService.create(new TaskModel({...task}))
		tasks.value.set(newTask.id, newTask)
	
		return newTask
	}

	async function updateTask(task: ITaskPartialWithId) {
		const oldTask = cloneDeep(tasks.value.get(task.id))

		if (!oldTask) return

		// we extend the task with potentially missing info
		const newTask: ITask = {
			...oldTask,
			...task,
		}

		// set in expectation that server update works
		tasks.value.set(newTask.id, newTask)

		try {	
			const updatedTask = await taskService.update(newTask)
			// update the task with possible changes from server
			tasks.value.set(updatedTask.id, updatedTask)
			success('Saved')
		} catch(e: any) {
			error('Something went wrong saving the task')
			// roll back changes
			tasks.value.set(task.id, oldTask)
		}
	}


	return {
		tasks,

		isLoading,
		loadTasks,

		addTask,
		updateTask,
	}
}