const _ = require('lodash');
const Joi = require('joi');
const Base = require('../../BaseJob');
const { TASK_PRIORITY } = require('../../../constants');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task DeletePvc',
};

class DeletePvcTask extends Base {
	async run(task) {
		this.logger.info('Running DeletePvc task');
		try {
			await this.kubernetesAPI.deletePvc(this.logger, task.spec.namespace, task.spec.name);
		} catch (err) {
			// we treat 404 as if the operation succeeded
			if (_.get(err, 'code') !== 404) {
				const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK}: ${err.message}`;
				this.logger.error(message);
				throw new Error(message);
			}
		}

		return Promise.resolve('OK');
	}

	async validate(task) {
		return Joi.validate(task, DeletePvcTask.validationSchema);
	}
}

DeletePvcTask.priority = TASK_PRIORITY.LOW;
DeletePvcTask.Errors = ERROR_MESSAGES;
DeletePvcTask.validationSchema = Joi.object().keys({
	spec: Joi.object().keys({
		namespace: Joi.string().required(),
		name: Joi.string().required()
	}).required(),
}).options({ stripUnknown: true });
module.exports       = DeletePvcTask;
