const _ = require('lodash');
const { create: createLogger } = require('../../services/Logger');
const ReportStatus = require('../ReportStatus');

jest.mock('./../../services/Logger');

describe('FetchTasksToExecute unit tests', () => {
	it('Should call Codefresh service to report the status', () => {
		const spy = jest.fn().mockResolvedValue('OK');
		const logger = createLogger();
		const task = new ReportStatus({
			reportStatus: spy,
		}, _.noop(), logger);
		const loggerMacher = expect.objectContaining({
			error: expect.any(Function),
			info: expect.any(Function),
			child: expect.any(Function),
		});
		return task.run()
			.then(() => {
				expect(spy).toHaveBeenCalled();
				expect(spy).toHaveBeenCalledWith(loggerMacher, { message: 'All good' });
			});
	});
});
