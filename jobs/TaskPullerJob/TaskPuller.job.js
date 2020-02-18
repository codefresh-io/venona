const Promise = require('bluebird');
const Chance = require('chance');
const _ = require('lodash');
const Base = require('../BaseJob');
const { TASK_PRIORITY } = require('../../constants');
const CreatePod = require('./tasks/CreatePod.task');
const DeletePod = require('./tasks/DeletePod.task');
const CreatePvc = require('./tasks/CreatePvc.task');
const DeletePvc = require('./tasks/DeletePvc.task');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run job TaskPuller, call to Codefresh rejected',
};

class TaskPullerJob extends Base {
	constructor(...args) {
		super(...args);
		
		this.typeToTaskMap = {
			'CreatePod': { executor: this._executeTask(CreatePod), priority: CreatePod.priority },
			'DeletePod': { executor: this._executeTask(DeletePod), priority: DeletePod.priority },
			'CreatePvc': { executor: this._executeTask(CreatePvc), priority: CreatePvc.priority },
			'DeletePvc': { executor: this._executeTask(DeletePvc), priority: DeletePvc.priority },
			'NOOP': { executor: _.noop, priority: TASK_PRIORITY.LOW },
		};
	}

	run() {
		return this.codefreshAPI.pullTasks(this.logger)
			.catch((err) => {
				const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			})
			.then(async (res = []) => {
				this.logger.info(`Got ${res.length} tasks`);

				const tasks = _.chain(res)
					.map((task) => {						
						const type = _.get(task, 'type');
						this.logger.info(`Got request to run task with type: ${type}`);
						const { executor, priority } = _.get(this.typeToTaskMap, type, this.typeToTaskMap.NOOP);
						return { task, priority, executor };
					})
					.filter(({ executor }) => executor !== _.noop)
					.sortBy(({ priority }) => priority)
					.value();

				// resolves each promise sequentially, in sorted order
				return await Promise.mapSeries(tasks, ({ task, executor }) => executor(task)); 
			});
	}

	_executeTask(Task) {
		return async (taskSpec) => {
			const logger = this.logger.child({
				task: Task.name,
				taskUid: new Chance().guid()
			});
			const task = new Task(this.codefreshAPI, this.runtimes, logger);
			return task.exec(taskSpec);
		};
	}

	validate() {
		return;
	}
}

TaskPullerJob.Errors = ERROR_MESSAGES;
module.exports = TaskPullerJob;
