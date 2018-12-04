const _ = require('lodash');
const { create: createLogger } = require('../../services/Logger');
const FetchTasksToExecute = require('../FetchTasksToExecute');
const StartWorkflow = require('./../StartWorkflow');

jest.mock('./../../services/Logger');
jest.mock('./../StartWorkflow');

describe('FetchTasksToExecute unit tests', () => {
	it('Should throw an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return expect(task.run()).rejects.toThrowError('Failed to run task FetchTasksToExecute, call to Codefresh rejected with message');
	});

	it('Should log an error when codefresh service call failed', () => {
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockRejectedValue(new Error('Failed')),
		}, _.noop(), logger);
		return task.run()
			.catch(() => {
				expect(logger.child.mock.results[1].value.error.mock.calls[0][0]).toMatch('Failed to run task FetchTasksToExecute, call to Codefresh rejected with message');
			});
	});

	it('Should pass logger to codefresh api service', () => {
		const spy = jest.fn().mockResolvedValue();
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: spy,
		}, _.noop(), logger);
		return task.run()
			.then(() => {
				expect(spy).toHaveBeenCalledWith(expect.objectContaining({
					error: expect.any(Function),
					info: expect.any(Function),
					child: expect.any(Function),
				}));
			});
	});

	it('Should map all results to tasks by types and execute them', () => {
		StartWorkflow.mockImplementationOnce(() => {
			return {
				run: jest.fn(() => {
					return {
						status: 'ok'
					};
				}),
			};
		});
		const tasks = [
			{
				type: 'StartWorkflow',
			}
		];
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.run()).resolves.toEqual([{ status: 'ok'}]);
	});

	it('Should not fail when unknown type task arrives', () => {
		const tasks = [
			{
				type: 'Fake-Task-Type',
			},
		];
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.run()).resolves.toEqual([]);
	});

	it('Should not fail when non-typed task arrives', () => {
		const tasks = [
			{
				prop: 'a'
			},
		];
		const logger = createLogger();
		const task = new FetchTasksToExecute({
			fetchTasksToExecute: jest.fn().mockResolvedValue(tasks),
		}, _.noop(), logger);
		return expect(task.run()).resolves.toEqual([]);
	});
});
