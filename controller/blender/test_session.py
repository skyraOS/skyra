import importlib.util
import json
import math
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
spec=importlib.util.spec_from_file_location('blender_session',Path(__file__).with_name('session.py'))
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
VERSION='a'*64
class ValidationTests(unittest.TestCase):
    def test_no_generic_code_or_paths(self):
        for r in [{'operation':'execute','code':'print(1)'},{'operation':'context','path':'/tmp/test'}, {'operation':'export_asset','expected_scene_version':VERSION,'output_path':'/tmp/test'}]:
            with self.assertRaises(m.Failure):m.validate(r)
    def test_transform_bounds_and_radians(self):
        r={'operation':'set_transform','object_id':'CoreMesh','expected_scene_version':VERSION,'rotation':[0,0.2,0]}
        self.assertEqual(m.validate(r),r)
        for v in [[0,0,math.nan],[0,0,11],[0,False,1],[1,2]]:
            with self.assertRaises(m.Failure):m.validate(dict(r,rotation=v))
    def test_material_allowlist(self):
        r={'operation':'set_material_parameter','material_id':'SkyraTeal','expected_scene_version':VERSION,'parameter':'roughness','value':0.4}
        self.assertEqual(m.validate(r),r)
        for p,v in [('code','os.remove'),('roughness',2),('base_color',[1,1,1])]:
            with self.assertRaises(m.Failure):m.validate(dict(r,parameter=p,value=v))
    def test_expected_version_required(self):
        with self.assertRaises(m.Failure):m.validate({'operation':'inspect_scene'})
    def test_malformed_field_types_report_invalid_request(self):
        for request in [{'operation':[]},{'operation':'inspect_scene','expected_scene_version':[]},
                        {'operation':'set_material_parameter','expected_scene_version':VERSION,'material_id':'SkyraTeal','parameter':[],'value':0}]:
            with self.assertRaises(m.Failure) as error:m.validate(request)
            self.assertEqual(error.exception.value['code'],'invalid_request')
    def test_conflict_never_dispatches_worker(self):
        with tempfile.TemporaryDirectory() as tmp:
            b=m.Bridge(None,Path(tmp)/'state',Path(tmp)/'output')
            b.binary=Path('/fake/blender')
            with patch.object(m.subprocess,'run') as run:
                with self.assertRaises(m.Failure) as error:b.handle({'operation':'inspect_scene','expected_scene_version':VERSION})
                self.assertEqual(error.exception.value['code'],'scene_version_conflict');run.assert_not_called()
    def test_profile_has_bounded_writes_and_no_network(self):
        sb=m.profile(Path('/Applications/Blender.app/Contents/MacOS/Blender'),Path('/approved/worker.py'),Path('/source/core.blend'),Path('/job/output'))
        self.assertIn('(deny default)',sb)
        self.assertNotIn('(allow network',sb)
        self.assertIn('(allow file-write* (subpath "/job/output") (literal "/dev/null"))',sb)
        self.assertNotIn(str(Path.home()),sb)
if __name__=='__main__':unittest.main()
