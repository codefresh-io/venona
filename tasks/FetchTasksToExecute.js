const Promise = require('bluebird');
const _ = require('lodash');
const Base = require('./BaseTask');
const StartWorkflow = require('./StartWorkflow');
const TerminateWorkflow = require('./TerminateWorkflow');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task FetchTasksToExecute, call to Codefresh rejected',
};

class FetchTasksToExecute extends Base {
	_executeTask(Task) {
		return (taskDef) => {
			const logger = this.logger.child({
				subTask: Task.name,
				workflow: taskDef.workflow,
				uid: this.uid,
			});
			const args = [this.codefreshAPI, this.kubernetesAPI, logger];
			return new Task(...args).run(taskDef);
		};
	}

	run() {
		this.logger.info('Running task FetchTasksToExecute');
		return this.codefreshAPI.fetchTasksToExecute(this.logger)
			.catch((err) => {
				const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK} with message: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			})
			.then((res = []) => {
				this.logger.info(`Got ${res.length} tasks`);
				const promises = _.chain(res)
					.map((task) => {
						const typeToTaskMap = {
							'StartWorkflow': this._executeTask(StartWorkflow),
							'FinishSystemWorkflow': this._executeTask(TerminateWorkflow),
						};
						const type = _.get(task, 'type');
						this.logger.info(`Got reqeust to run task with type: ${type}`);
						const fn = typeToTaskMap[type] || _.noop;
						return fn(task);
					})
					.compact()
					.flattenDeep()
					.value();
				return Promise.all(promises);
			});
	}
}
FetchTasksToExecute.Errors = ERROR_MESSAGES;
module.exports = FetchTasksToExecute;
