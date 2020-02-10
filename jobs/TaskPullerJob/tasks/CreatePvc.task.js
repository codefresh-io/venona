const Joi = require('joi');
const Base = require('../../BaseJob');
const _ = require('lodash');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task CreatePvc',
};

class CreatePvcTask extends Base {
	async run(task) {
		this.logger.info('Running CreatePvc task');
		try {
			const pvc = await this.kubernetesAPI.createPvc(this.logger, task.spec, _.get(task, 'metadata.reName'));
			return pvc;
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK}: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}

	async validate(task) {
		return Joi.validate(task, CreatePvcTask.validationSchema);
	}
}
CreatePvcTask.Errors = ERROR_MESSAGES;
CreatePvcTask.validationSchema = Joi.object().keys({
	spec: Joi.object().required(),
}).options({ stripUnknown: true });
module.exports       = CreatePvcTask;
