# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import datetime
import logging
import time
import urllib2

from cuckoo.core.database import Database
from cuckoo.massurl import db
from ektesting.abstracts import FuncTest
from ektesting.helpers import rand_string, settings

log = logging.getLogger(__name__)

class RigEKSubmit(FuncTest):
    name = "RigEKSubmit"
    description = "Test a submission of a RigEK exploit URL"

    def init(self):
        Database().connect()
        self.groupname = rand_string(16)
        self.group_id = None

    def start(self):
        tags = {t.name:t.id for t in db.db.list_tags()}
        profile_tags = ["windows7", "ie11", "flash2000228"]
        for t in profile_tags:
            if t not in tags:
                return self.markfail(
                    "Missing machine tag '%s'. There much be a machine with the"
                    " machine tags %s. This is required to run the EK detection"
                    " test " % (t,profile_tags),
                    fix="Create/add a Windows 7 VM with Internet Explorer 11"
                        " and Flash 20.0.0.228 and add the tags: %s to this"
                        " machine in the machinery config." % profile_tags
                )

        self.group_id = db.add_group(self.groupname, rand_string(20))
        db.update_settings_group(
            group_id=self.group_id, batch_time=60, threshold=10, batch_size=1
        )

        rigekurl = "%s/rigekexploit.html" % settings.webserver
        try:
            urllib2.urlopen(rigekurl).read()
        except Exception:
            return self.markfail(
                "Failed to perform GET request to URL '%s'" % rigekurl,
                fix="Use the mini webserver in the 'scripts' folder to serve "
                    "the files in the 'data' folder. Make sure the current "
                    "server is allowed to connect to it."
            )

        db.mass_group_add([rigekurl], group_id=self.group_id)
        profile_id = db.add_profile(
            name=rand_string(12), browser="ie", route="internet",
            tags=[tid for n, tid in tags.iteritems() if n in profile_tags]
        )
        db.update_profile_group([profile_id], self.group_id)
        db.set_schedule_next(
            group_id=self.group_id,
            next_datetime=datetime.datetime.utcnow()
        )

        log.debug(
            "Scheduled group. Waiting at least 50 seconds before checking"
        )
        time.sleep(50)
        while True:
            group = db.find_group(group_id=self.group_id)
            if group.status == "completed":
                break
            log.debug("Group status still '%s'. Waiting..", group.status)
            time.sleep(5)

        alerts = db.list_alerts(level=3, url_group_name=self.groupname)
        if alerts:
            return self.markpass()
        else:
            return self.markfail(
                "Group has completed, but level 3 alerts exist. One level 3 "
                "alert was expected.",
                fix="Make sure the analysis VM has internet explorer 11 with "
                    "Flash 20.0.0.228 installed using VMCloak. "
                    "Verify the analysis logs to see if Onemon was loaded. "
                    "It might happen an exploit does not trigger. It is "
                    "advised to run this test at least more than once in case "
                    "of a fail"
            )

    def cleanup(self):
        if self.group_id:
            db.delete_group(group_id=self.group_id)
            db.delete_alert(group_name=self.groupname)
