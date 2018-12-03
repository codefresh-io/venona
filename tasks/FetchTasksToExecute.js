const Promise = require('bluebird');
const _ = require('lodash');
const Base = require('./BaseTask');
const CreateKubernetesPod = require('./CreateKubernetesPod');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task FetchTasksToExecute, call to Codefresh rejected',
};

class FetchTasksToExecute extends Base {
	_createKubernetesPod(id, spec) {
		this.logger.info(`Workflow: ${id}`);
		const logger = this.logger.child({
			subTask: CreateKubernetesPod.name,
			workflow: id,
			uid: this.uid,
			name: _.get(spec, 'metadata.name'),
		});
		const args = [this.codefreshAPI, this.kubernetesAPI, logger];
		return new CreateKubernetesPod(...args).run(spec);
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
					.map(({ dockerDaemon, runtime, workflow }) => Promise.all([
						this._createKubernetesPod(workflow, dockerDaemon),
						this._createKubernetesPod(workflow, runtime),
					]))
					.flattenDeep()
					.value();
				return Promise.all(promises);
			});
	}
}
FetchTasksToExecute.Errors = ERROR_MESSAGES;
module.exports = FetchTasksToExecute;
