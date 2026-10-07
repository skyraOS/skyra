"""Trusted Blender fixture builder. Run only with a pinned Blender in a bounded worker.

This is a developer pilot, not a model-facing generic script execution tool.
"""
import argparse
import hashlib
import json
import math
import sys
from pathlib import Path
import bpy


def material(name, color, metallic=0.0, roughness=0.55):
    m = bpy.data.materials.new(name)
    m.diffuse_color = (*color, 1)
    m.use_nodes = True
    node = m.node_tree.nodes.get('Principled BSDF')
    node.inputs['Base Color'].default_value = (*color, 1)
    node.inputs['Metallic'].default_value = metallic
    node.inputs['Roughness'].default_value = roughness
    return m


def build(out):
    bpy.ops.wm.read_factory_settings(use_empty=True)
    bpy.context.preferences.filepaths.temporary_directory = str(out)
    collection = bpy.data.collections.new('SKYRA_EXPORT')
    bpy.context.scene.collection.children.link(collection)
    root = bpy.data.objects.new('SkyraCoreRoot', None)
    collection.objects.link(root)
    teal = material('SkyraTeal', (0.15, 0.35, 0.39), 0.4)
    accent = material('SkyraAccent', (0.49, 0.77, 0.80), 0.3)
    for name, radius, major in [('CoreMesh', 0.4, None), ('ArcInner', 0.015, 0.53), ('ArcOuter', 0.012, 0.64)]:
        if major is None:
            bpy.ops.mesh.primitive_ico_sphere_add(subdivisions=3, radius=radius)
        else:
            bpy.ops.mesh.primitive_torus_add(major_segments=64, minor_segments=8,
                                           major_radius=major, minor_radius=radius)
        obj = bpy.context.object
        obj.name = name
        obj.parent = root
        if major is not None:
            obj.rotation_euler = (math.radians(35 if name == 'ArcInner' else -30), math.radians(15), 0)
        for c in list(obj.users_collection):
            c.objects.unlink(obj)
        collection.objects.link(obj)
        obj.data.materials.append(teal if major is None else accent)
        for p in obj.data.polygons:
            p.use_smooth = True
    bpy.context.scene.unit_settings.system = 'METRIC'
    # Blender Z-up source; USDZ exports Y-up explicitly, avoiding adapter double rotation.
    bpy.ops.wm.save_as_mainfile(filepath=str(out / 'core.blend'))
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.wm.usd_export(filepath=str(out / 'core.usdz'), selected_objects_only=True,
                          export_materials=True, export_textures_mode='NEW',
                          convert_orientation=True, export_global_forward_selection='NEGATIVE_Z',
                          export_global_up_selection='Y')
    bpy.ops.export_scene.gltf(filepath=str(out / 'core.glb'), export_format='GLB', use_selection=True)
    report = {'blender': bpy.app.version_string, 'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
              'source_up': 'Z', 'runtime_up': 'Y', 'nodes': ['SkyraCoreRoot', 'CoreMesh', 'ArcInner', 'ArcOuter'],
              'triangles': sum(sum(len(p.vertices)-2 for p in o.data.polygons) for o in collection.objects if o.type=='MESH'),
              'materials': 2, 'files': {}}
    for name in ['core.blend', 'core.usdz', 'core.glb']:
        p = out / name
        report['files'][name] = {'sha256': hashlib.sha256(p.read_bytes()).hexdigest(), 'bytes': p.stat().st_size}
    # Axis fixture is deliberately asymmetric, with named RGB arrows.
    bpy.ops.wm.read_factory_settings(use_empty=True)
    for name, color, location, scale in [('AxisX',(1,0,0),(0.4,0,0),(0.4,0.04,0.04)),
                                        ('AxisY',(0,1,0),(0,0.3,0),(0.04,0.3,0.04)),
                                        ('AxisZ',(0,0,1),(0,0,0.2),(0.04,0.04,0.2))]:
        bpy.ops.mesh.primitive_cube_add(size=2, location=location)
        obj=bpy.context.object;obj.name=name;obj.scale=scale
        obj.data.materials.append(material(name, color))
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.wm.usd_export(filepath=str(out/'axis.usdz'), selected_objects_only=True, export_materials=True,
                          convert_orientation=True, export_global_forward_selection='NEGATIVE_Z', export_global_up_selection='Y')
    p=out/'axis.usdz';report['files'][p.name]={'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'bytes':p.stat().st_size}
    (out/'build-report.json').write_text(json.dumps(report, indent=2)+'\n')
    print(json.dumps(report))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--out', required=True)
    args = parser.parse_args(sys.argv[sys.argv.index('--')+1:])
    out = Path(args.out).resolve(strict=True)
    if not out.is_dir():
        raise ValueError('Output must be an existing granted job directory')
    if any((out/name).exists() for name in ['core.blend','core.usdz','core.glb','axis.usdz','build-report.json']):
        raise ValueError('Refusing to overwrite an existing pilot job')
    build(out)
