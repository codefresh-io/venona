const Joi = require('joi');
const Base = require('../../BaseJob');
const _ = require('lodash');
const { TASK_PRIORITY } = require('../../../constants');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task CreatePvc',
};

class CreatePvcTask extends Base {
	async run(task) {
		this.logger.info('Running CreatePvc task');
		try {
			const service = await this.getKubernetesService(_.get(task, 'metadata.reName'));
			const pvc = await service.createPvc(this.logger, task.spec);
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

CreatePvcTask.priority = TASK_PRIORITY.HIGH;
CreatePvcTask.Errors = ERROR_MESSAGES;
CreatePvcTask.validationSchema = Joi.object().keys({
	spec: Joi.object().required(),
}).options({ stripUnknown: true });

module.exports       = CreatePvcTask;
