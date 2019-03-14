# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import logging
import time

from cuckoo.core.realtime import EventClient
from ektesting.abstracts import FuncTest

log = logging.getLogger(__name__)

class UseEventServer(FuncTest):
    name = "Eventserver test"
    description = "Test the eventserver connection and test if events" \
                  " can be sent and received"

    order = 999
    stop_on_fail = True

    def init(self):
        self.ev = EventClient(self.settings.event_ip, self.settings.event_port)
        self.eventtest = False

    def _handle_event(self, m):
        if m["body"].get("data") == "test":
            self.eventtest = True

    def start(self):
        if not self.ev.start(2):
            return self.markfail(
                "Connecting to eventserver failed",
                fix="Verify if the event server is online and this machine's"
                    " IP is whitelisted in the cuckoo.conf"
            )

        self.ev.subscribe(self._handle_event, "ekhuntingtest")
        ev2 = EventClient(self.settings.event_ip, self.settings.event_port)
        ev2.start(2)
        ev2.send_event("ekhuntingtest", {"data": "test"})
        time.sleep(3)
        if not self.eventtest:
            return self.markfail(
                "Testing event not received after sending",
            )

        self.eventtest = False
        self.ev.unsubscribe(self._handle_event, "ekhuntingtest")
        ev2.send_event("ekhuntingtest", {"data": "test"})
        time.sleep(3)
        if self.eventtest:
            return self.markfail(
                "Unsubscribe did not work. Still received event"
            )
        return self.markpass()

    def cleanup(self):
        self.ev.disconnect()
        self.ev.stop()
