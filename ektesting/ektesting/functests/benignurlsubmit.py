# Copyright (C) 2019 Hatching B.V.
# All rights reserved.

import datetime
import logging
import time

from cuckoo.core.database import Database
from cuckoo.massurl import db
from cuckoo.massurl.urldiary import URLDiaries
from cuckoo.common.objects import URL

from ektesting.abstracts import FuncTest
from ektesting.helpers import rand_string, settings

log = logging.getLogger(__name__)

class BenignURLSubmit(FuncTest):
    name = "BenignURLSubmit"
    description = "Submit multiple benign URLs and verify there are " \
                  "no detections"

    def init(self):
        Database().connect()
        URLDiaries.init()
        self.groupname = rand_string(16)
        self.group_id = None

    def start(self):
        log.debug("Using group with name '%s'", self.groupname)
        tags = {t.name:t.id for t in db.db.list_tags()}
        profile_tags = ["windows7", "ie11"]
        for t in profile_tags:
            if t not in tags:
                return self.markfail(
                    "Missing machine tag '%s'. There much be a machine with the"
                    " machine tags %s. This is required to run the EK detection"
                    " test " % (t,profile_tags),
                    fix="Create/add a Windows 7 VM with Internet Explorer 11"
                        "  and add the tags: %s to this machine in the"
                        " machinery config." % profile_tags
                )

        self.group_id = db.add_group(self.groupname, rand_string(20))
        db.update_settings_group(
            group_id=self.group_id, batch_time=30, threshold=10, batch_size=5
        )

        urls = [
            "http://facebook.com", "http://baidu.com", "http://wikipedia.org",
            "http://qq.com", "http://taobao.com"
        ]
        db.mass_group_add(urls, group_id=self.group_id)
        profile_id = db.add_profile(
            name=rand_string(12), browser="ie", route="internet",
            tags=[tid for n, tid in tags.iteritems() if n in profile_tags]
        )
        db.update_profile_group([profile_id], self.group_id)
        db.set_schedule_next(
            group_id=self.group_id,
            next_datetime=datetime.datetime.utcnow()
        )

        beforefinish = int(time.time() * 1000)
        log.debug(
            "Scheduled group. Waiting at least 30 seconds before checking"
        )
        time.sleep(30)
        while True:
            group = db.find_group(group_id=self.group_id)
            if group.status == "completed":
                break

            log.debug("Group status still '%s'. Waiting..", group.status)
            time.sleep(5)

        requests_extracted = False
        for url in urls:
            url = URL(url)
            diary = URLDiaries.get_latest_diary(
                url.get_sha256(), return_fields=[
                    "datetime", "requested_urls", "signatures"
                ]
            )
            if not diary or not diary.get("datetime", 0) > beforefinish:
                return self.markfail(
                    "No new URL diary was created for %s" % url
                )

            if len(diary.get("requested_urls")):
                requests_extracted = True

            if len(diary.get("signatures")):
                return self.markfail(
                    "One ore more realtime signature were triggered. This "
                    "should not happen for these submitted URLs '%s'. Check "
                    "the result." % urls
                )

        if not requests_extracted:
            return self.markfail(
                "No HTTP requests were extracted for any of the analyzed URLs",
                fix="Is internet routing enabled and Cuckoo rooter running? "
                    "Verify if onemon.pb exists in the logs directory to "
                    "see if onemon collected any events."
            )

        alerts = db.list_alerts(level=3, url_group_name=self.groupname)
        if alerts:
            return self.markfail(
                "One or more level 3 alerts was triggered. "
                "No level 3 alerts should have been sent."
            )

        return self.markpass()

    def cleanup(self):
        if self.group_id:
            db.delete_group(group_id=self.group_id)
            db.delete_alert(group_name=self.groupname)
