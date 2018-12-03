const _ = require('lodash');
const BaseTask = require('./../BaseTask');
const { create: createLogger } = require('./../../services/Logger');

jest.mock('./../../services/Logger');

describe('BaseTask unit tests', () => {
	it('Should construct', () => {
		const task = new BaseTask(_.noop(), _.noop(), createLogger());
		expect(Object.keys(task).sort()).toEqual(['codefreshAPI', 'kubernetesAPI', 'logger'].sort());
	});
});
