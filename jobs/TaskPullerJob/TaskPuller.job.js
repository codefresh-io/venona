const Promise = require('bluebird');
const Chance = require('chance');
const _ = require('lodash');
const Base = require('../BaseJob');
const CreatePod = require('./tasks/CreatePod.task');
const DeletePod = require('./tasks/DeletePod.task');
const CreatePvc = require('./tasks/CreatePvc.task');
const DeletePvc = require('./tasks/DeletePvc.task');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run job TaskPuller, call to Codefresh rejected',
};

class TaskPullerJob extends Base {
	run() {
		return this.codefreshAPI.pullTasks(this.logger)
			.catch((err) => {
				const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			})
			.then((res = []) => {
				if (_.isEmpty(res)) {
					this.logger.infoVerbose('Got 0 tasks');
				} else {
					this.logger.info(`Got ${res.length} new tasks`);
				}
				const promises = _.chain(res)
					.map((task) => {
						// TODO auto load all tasks
						const typeToTaskMap = {
							'CreatePod': this._executeTask(CreatePod),
							'DeletePod': this._executeTask(DeletePod),
							'CreatePvc': this._executeTask(CreatePvc),
							'DeletePvc': this._executeTask(DeletePvc),
						};
						const type = _.get(task, 'type');
						this.logger.info(`Got request to run task with type: ${type}`);
						const fn = typeToTaskMap[type] || _.noop;
						return fn(task);
					})
					.compact()
					.flattenDeep()
					.value();
				return Promise.all(promises);
			});
	}

	_executeTask(Task) {
		return async (taskSpec) => {
			const logger = this.logger.child({
				task: Task.name,
				taskUid: new Chance().guid()
			});
			const task = new Task(this.codefreshAPI, this.kubernetesAPI, logger);
			return task.exec(taskSpec);
		};
	}

	validate() {
		return;
	}
}
TaskPullerJob.Errors = ERROR_MESSAGES;
module.exports = TaskPullerJob;
