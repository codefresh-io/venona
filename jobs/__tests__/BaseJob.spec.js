const _ = require('lodash');
const BaseJob = require('./../BaseJob');
const { create: createLogger } = require('../../services/Logger');

jest.mock('./../../services/Logger');

describe('BaseJob unit tests', () => {

	describe('positive', () => {
		it('Should construct', () => {
			const task = new BaseJob(_.noop(), _.noop(), createLogger());
			expect(Object.keys(task).sort()).toEqual(['codefreshAPI', 'runtimes', 'logger'].sort());
		});

		it.skip('Should throw an error in case the requested runtime is not set on this.runtimes', () => {});
	});

});
