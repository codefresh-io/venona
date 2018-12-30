const _ = require('lodash');
const BaseJob = require('./../BaseJob');
const { create: createLogger } = require('../../services/Logger');

jest.mock('./../../services/Logger');

describe('BaseJob unit tests', () => {

	describe('positive', () => {
		it('Should construct', () => {
			const task = new BaseJob(_.noop(), _.noop(), createLogger());
			expect(Object.keys(task).sort()).toEqual(['codefreshAPI', 'kubernetesAPI', 'logger'].sort());
		});
	});

	describe('negative', () => {

	});

});
