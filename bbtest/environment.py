#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper
from mocks.bondster.server import BondsterMock
from mocks.vault.server import VaultMock
from mocks.ledger.server import LedgerMock
from helpers.statsd import StatsdHelper


def after_feature(context, feature):
  context.unit.collect_logs()


def before_all(context):
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper(context)
  context.bondster = BondsterMock(context)
  context.ledger = LedgerMock(context)
  context.vault = VaultMock(context)
  context.statsd = StatsdHelper()
  context.statsd.start()
  context.bondster.start()
  context.ledger.start()
  context.vault.start()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.bondster.stop()
  context.ledger.stop()
  context.vault.stop()
  context.unit.teardown()
  context.zmq.stop()
  context.statsd.stop()
