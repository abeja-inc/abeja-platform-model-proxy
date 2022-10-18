import collections
import importlib
import json
import logging
import os
import subprocess
import signal
import sys
import uuid
from collections import deque
from datetime import (datetime, timezone)
from pathlib import Path
from traceback import format_tb
from typing import Callable, List, Optional, Tuple

CONFIGURE_BASE_DIRECTORY = '.abeja'
WORKING_DIRECTORY_SEARCH_DEPTH = 3
DEFAULT_USER_LOG_FORMAT = '%(levelname)s: %(message)s'

signal.signal(signal.SIGINT, signal.SIG_DFL)


class SubprocessException(Exception):
    def __init__(self, e: subprocess.CalledProcessError):
        super(SubprocessException, self).__init__(str(e))
        self.exc = e

    def get_output(self) -> bytes:
        return self.exc.output

    def get_returncode(self) -> int:
        return self.exc.returncode


class ModelSetupException(Exception):
    pass


class InvalidDatasetAliasError(Exception):
    pass

"""
logging module from
"""

LogRecordNamedtuple = collections.namedtuple(
    'LogRecordNamedtuple', 'log_id log_level timestamp source requester_id message exc_info'
)


class JsonLogFormatter(logging.Formatter):

    def format(self, log_record):
        """
        Format the record as json style string.

        :param log_record: (LogRecord) An event being logged.
        :return: (str) json style log string.
        """
        # Get log record.
        log_id = str(uuid.uuid4())
        log_level = log_record.levelname
        timestamp = datetime.fromtimestamp(log_record.created, tz=timezone.utc).isoformat()
        log_category = log_record.log_category if hasattr(log_record, 'log_category') else '{}.{}.{}'.format(
            log_record.module,
            log_record.funcName,
            log_record.lineno
        )
        source = '{}:{}'.format(log_record.name, log_category)
        requester_id = log_record.requester_id if hasattr(log_record, 'requester_id') else '-'
        message = log_record.getMessage()
        if log_record.exc_info:
            error_type, error, tb = log_record.exc_info
            exc_info = {
                'type': str(error_type),
                'error': str(error),
                'traceback': format_tb(tb)
            }
        else:
            exc_info = None
        log_record_namedtuple = LogRecordNamedtuple(log_id, log_level, timestamp, source, requester_id,
                                                    message, exc_info)
        log_record_dict = log_record_namedtuple._asdict()
        # Add optional keys.
        if hasattr(log_record, 'options'):
            reserved_key = log_record_dict.keys()
            for key, value in log_record.options.items():
                if key not in reserved_key:
                    log_record_dict[key] = value
        # Return json style log string.
        return json.dumps(log_record_dict)


def _build_logger(service_name: str, use_json_log_formatter: bool, loglevel) -> logging.Logger:
    logger = logging.getLogger(service_name)
    logger.setLevel(loglevel)
    handler = logging.StreamHandler(sys.stdout)
    if use_json_log_formatter:
        handler.setFormatter(JsonLogFormatter())
    else:
        handler.setFormatter(logging.Formatter(DEFAULT_USER_LOG_FORMAT))
    handler.setLevel(loglevel)
    logger.addHandler(handler)
    logger.propagate = False
    return logger


stage = os.environ.get('STAGE')
if stage == 'prod':
    loglevel = logging.INFO
else:
    loglevel = logging.DEBUG

# The logger for runtime code
runtime_logger = _build_logger('runtime', use_json_log_formatter=True, loglevel=loglevel)

# The logger for user code
user_logger = _build_logger('user', use_json_log_formatter=False, loglevel=loglevel)


"""
search working dir
"""

def _find_working_dir() -> Optional[str]:
    """
    Let the working directory be the path where the .abeja directory resides.

    """
    base_path = Path(os.curdir)
    base_depth = str(base_path.resolve()).count(os.sep)
    q: deque = deque([])
    q.append(base_path)
    while len(q) > 0:
        p = q.popleft()
        for child in p.iterdir():
            current_depth = str(child.resolve()).count(os.sep)
            if child.is_dir():
                if child.name == CONFIGURE_BASE_DIRECTORY:
                    return str(child.parent.resolve())
                else:
                    if (current_depth - base_depth) > WORKING_DIRECTORY_SEARCH_DEPTH:
                        continue
                    q.append(child)
    return None


"""
module import libs
"""

def _run(command: List[str]) -> Tuple[int, bytes]:
    try:
        process = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=True)
        return process.returncode, process.stdout
    except subprocess.CalledProcessError as e:
        raise SubprocessException(e)


def _install_packages() -> None:
    if os.path.exists('Pipfile'):
        user_logger.info('start installing packages from Pipfile')
        _install_from_pipfile()
        user_logger.info('packages are installed from Pipfile')
    elif os.path.exists('requirements.txt'):
        user_logger.info('start installing packages from requirements.txt')
        _install_from_requirements_txt()
        user_logger.info('packages are installed from requirements.txt')
    else:
        user_logger.info('requirements.txt/Pipfile not found, skip installing dependencies process')


def _install_from_pipfile() -> None:
    commands = ['pipenv', 'install', '--system']
    if not os.path.exists('Pipfile.lock'):
        commands.append('--skip-lock')
    _run(commands)


def _install_from_requirements_txt(requirements_path: str = 'requirements.txt') -> None:
    """installs python packages described in requirements
    Args:
        requirements_path:
    Returns:
        ``True`` if pip install succeeded.
        ``False`` if requirements.txt not found.
    Raises:
        CalledProcessError - failed to `pip install`
    """
    _run(['pip', 'install', '-q', '-r', requirements_path])


def _import_train(handler: str) -> Callable:
    """loads user model identified by handler
    Args:
        handler: path to the handler of user definied model
    Returns:
        function object of user defined model
    Raises:
        ImportError: failed to import user definied model
    """

    if handler.count(':') != 1:
        raise ImportError(
            'Possibility, your HANDLER[{}] parameter is wrong. '
            'handler needs one ":" separator. The format is <module>:<function>.'
            .format(handler))

    module_name, func_name = handler.split(':', 1)
    runtime_logger.debug(f'HANDLER: module_name: {module_name}, func_name: {func_name}')

    # Temporarily modify sys.path to include current working directory.
    if '' not in sys.path:
        sys.path.insert(0, '')

    try:
        model = getattr(importlib.import_module(module_name), func_name)
    except ImportError as e:
        raise ImportError(
            "Couldn't import a module named [{}] specified in HANDLER[{}]. "
            "Possibly, your HANDLER parameter or packaging (zip/tar) file structure is wrong. {}"
            .format(module_name, handler, e.__str__()))
    except AttributeError:
        raise ImportError(
            "Couldn't import a function named [{}] specified in HANDLER[{}]. "
            "Possibly, your HANDLER parameter or packaging (zip/tar) file structure is wrong."
            .format(func_name, handler))

    return model


def setup_train(handler: str) -> Callable:
    working_dir = _find_working_dir()
    if working_dir:
        user_logger.debug('found .abeja directory in {}.'.format(working_dir))
        try:
            os.chdir(working_dir)
        except Exception as e:
            raise ModelSetupException(str(e))

    try:
        _install_packages()
    except SubprocessException as e:
        user_logger.error(
            'error while installing from requirements.txt/Pipfile:\n' + e.get_output().decode('utf-8'))
        tb = sys.exc_info()[2]
        raise ModelSetupException(e).with_traceback(tb)

    try:
        user_logger.info("Loading user model handler with {}...".format(handler))
        model = _import_train(handler)
        return model
    except ImportError as e:
        user_logger.exception('error while importing user model')
        raise ModelSetupException('Exception occurred while loading handler function: {}'.format(str(e)))


def normalize_dataset_alias() -> dict:
    datasets_str = os.environ.get('DATASETS', '{}')
    runtime_logger.debug(f'dataset alias: {datasets_str}')
    datasets_json = json.loads(datasets_str)
    normalized_datasets = {}
    for key, value in datasets_json.items():
        if not isinstance(key, str):
            raise InvalidDatasetAliasError(f'{key} is invalid dataset alias key')
        try:
            value = int(value) if isinstance(value, str) else value
        except ValueError:
            raise InvalidDatasetAliasError(f'value for {key} is invalid dataset identifier, '
                                           'dataset identifier should be integer')
        if not isinstance(value, int):
            raise InvalidDatasetAliasError(f'value for {key} is invalid dataset identifier, '
                                           'dataset identifier should be integer')
        normalized_datasets[key] = str(value)
    return normalized_datasets


def main():
    handler = os.environ.get('HANDLER')
    if handler is None:
        user_logger.error("HANDLER is needed")
        sys.exit(1)
    try:
        train = setup_train(handler)
    except ModelSetupException:
        user_logger.warning("failed to setup training", exc_info=True)
        sys.exit(1)

    try:
        context = json.loads(os.environ.get('CONTEXT', '{}'))
    except json.decoder.JSONDecodeError:
        user_logger.exception('Exception occurred while loading handler context')
        sys.exit(1)

    try:
        datasets = normalize_dataset_alias()
    except (InvalidDatasetAliasError, json.decoder.JSONDecodeError):
        user_logger.exception('Invalid dataset alias')
        sys.exit(1)
    context['datasets'] = datasets
    runtime_logger.debug(f'context = {context}')

    try:
        train(context)
    except Exception as e:
        user_logger.warning("unexpected error occured in user train-code", exc_info=True)
        sys.exit(1)

    sys.exit(0)


if __name__ == "__main__":
    main()

