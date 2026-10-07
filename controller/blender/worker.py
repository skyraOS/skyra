"""Fixed implementation of typed Blender operations; never evaluates request text."""
import hashlib
import json
import math
import sys
from pathlib import Path
import bpy


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def summary():
    objects=[]
    for obj in bpy.context.scene.objects:
        objects.append({'id':obj.name,'type':obj.type,'translation':list(obj.location),
                        'rotation':list(obj.rotation_euler),'scale':list(obj.scale),
                        'triangles':sum(len(p.vertices)-2 for p in obj.data.polygons) if obj.type=='MESH' else 0})
    return {'objects':objects,'materials':[m.name for m in bpy.data.materials],
            'triangles':sum(o['triangles'] for o in objects),'units':bpy.context.scene.unit_settings.system}


def main(config_path):
    if bpy.app.version < (5,1,0):raise ValueError('unsupported_version')
    config=json.loads(Path(config_path).read_text())
    request=config['request'];out=Path(config['output']);source=Path(config['source'])
    bpy.context.preferences.filepaths.temporary_directory=str(out)
    bpy.ops.wm.open_mainfile(filepath=str(source),use_scripts=False)
    bpy.context.preferences.filepaths.temporary_directory=str(out)
    before=summary();op=request['operation'];change=None
    if op=='inspect_object':
        obj=next((o for o in before['objects'] if o['id']==request['object_id']),None)
        if obj is None:raise ValueError('object_not_found')
        result=obj
    elif op=='set_transform':
        obj=bpy.data.objects.get(request['object_id'])
        if obj is None or obj not in bpy.context.scene.objects.values():raise ValueError('object_not_found')
        for field,property in [('translation','location'),('rotation','rotation_euler'),('scale','scale')]:
            if field in request:setattr(obj,property,request[field])
        change={'object_id':obj.name,'before':next(o for o in before['objects'] if o['id']==obj.name)}
        result=summary()
    elif op=='set_material_parameter':
        mat=bpy.data.materials.get(request['material_id'])
        if mat is None:raise ValueError('material_not_found')
        node=mat.node_tree.nodes.get('Principled BSDF') if mat.node_tree else None
        if node is None:raise ValueError('unsupported_material')
        field={'roughness':'Roughness','metallic':'Metallic','base_color':'Base Color'}[request['parameter']]
        old=node.inputs[field].default_value
        change={'material_id':mat.name,'parameter':request['parameter'],'before':list(old) if hasattr(old,'__len__') else old,'after':request['value']}
        node.inputs[field].default_value=request['value']
        result=summary()
    else:result=before
    after=summary()
    if after['triangles']>20000 or len(after['materials'])>4 or len(after['objects'])>32:raise ValueError('asset_budget_exceeded')
    if op in ('set_transform','set_material_parameter'):
        if change.get('object_id'):change['after']=next(o for o in after['objects'] if o['id']==change['object_id'])
        bpy.ops.wm.save_as_mainfile(filepath=str(out/'scene.blend'))
    if op=='export_asset':
        collection=bpy.data.collections.get('SKYRA_EXPORT')
        if collection is None:raise ValueError('missing_export_collection')
        bpy.ops.object.select_all(action='DESELECT')
        for obj in collection.all_objects:obj.select_set(True)
        bpy.ops.wm.usd_export(filepath=str(out/'asset.usdz'),selected_objects_only=True,export_materials=True,
                              export_textures_mode='NEW',convert_orientation=True,
                              export_global_forward_selection='NEGATIVE_Z',export_global_up_selection='Y',
                              export_lights=False,export_cameras=False,export_animation=False)
        bpy.ops.export_scene.gltf(filepath=str(out/'asset.glb'),export_format='GLB',use_selection=True)
    if op=='render_preview':
        from mathutils import Vector
        scene=bpy.context.scene
        scene.render.engine='CYCLES';scene.cycles.device='CPU';scene.cycles.samples=16
        scene.render.resolution_x=384;scene.render.resolution_y=384;scene.render.resolution_percentage=100
        scene.render.film_transparent=True
        bpy.ops.object.camera_add(location=(0,-2.7,1.2))
        cam=bpy.context.object;cam.rotation_euler=(-cam.location).to_track_quat('-Z','Y').to_euler();scene.camera=cam
        bpy.ops.object.light_add(type='AREA',location=(2,-3,4))
        light=bpy.context.object;light.data.energy=400;light.data.shape='DISK';light.data.size=4
        light.rotation_euler=(-light.location).to_track_quat('-Z','Y').to_euler()
        scene.render.image_settings.file_format='PNG';scene.render.filepath=str(out/'preview.png')
        bpy.ops.render.render(write_still=True)
    files={}
    for name in ['scene.blend','asset.usdz','asset.glb','preview.png']:
        p=out/name
        if p.is_file():files[name]={'sha256':sha(p),'size':p.stat().st_size}
    report={'version':1,'operation':'blender.tools','requested_operation':op,'status':'succeeded',
            'source_version':sha(source),'blender_version':bpy.app.version_string,'result':result,
            'change':change,'files':files}
    (out/'result.json').write_text(json.dumps(report,indent=2)+'\n')


if __name__=='__main__':
    main(sys.argv[sys.argv.index('--')+1])
