// +build !ignore_autogenerated

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by conversion-gen. Do not edit it manually!

package v1

import (
	api "k8s.io/kubernetes/pkg/api"
	unversioned "k8s.io/kubernetes/pkg/api/unversioned"
	autoscaling "k8s.io/kubernetes/pkg/apis/autoscaling"
	extensions "k8s.io/kubernetes/pkg/apis/extensions"
	conversion "k8s.io/kubernetes/pkg/conversion"
	reflect "reflect"
)

func init() {
	if err := api.Scheme.AddGeneratedConversionFuncs(
		Convert_v1_HorizontalPodAutoscaler_To_extensions_HorizontalPodAutoscaler,
		Convert_extensions_HorizontalPodAutoscaler_To_v1_HorizontalPodAutoscaler,
		Convert_v1_HorizontalPodAutoscalerList_To_extensions_HorizontalPodAutoscalerList,
		Convert_extensions_HorizontalPodAutoscalerList_To_v1_HorizontalPodAutoscalerList,
		Convert_v1_HorizontalPodAutoscalerSpec_To_extensions_HorizontalPodAutoscalerSpec,
		Convert_extensions_HorizontalPodAutoscalerSpec_To_v1_HorizontalPodAutoscalerSpec,
		Convert_v1_HorizontalPodAutoscalerStatus_To_extensions_HorizontalPodAutoscalerStatus,
		Convert_extensions_HorizontalPodAutoscalerStatus_To_v1_HorizontalPodAutoscalerStatus,
		Convert_v1_Scale_To_autoscaling_Scale,
		Convert_autoscaling_Scale_To_v1_Scale,
		Convert_v1_ScaleSpec_To_autoscaling_ScaleSpec,
		Convert_autoscaling_ScaleSpec_To_v1_ScaleSpec,
		Convert_v1_ScaleStatus_To_autoscaling_ScaleStatus,
		Convert_autoscaling_ScaleStatus_To_v1_ScaleStatus,
	); err != nil {
		// if one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}

func autoConvert_v1_HorizontalPodAutoscaler_To_extensions_HorizontalPodAutoscaler(in *HorizontalPodAutoscaler, out *extensions.HorizontalPodAutoscaler, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*HorizontalPodAutoscaler))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	// TODO: Inefficient conversion - can we improve it?
	if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
		return err
	}
	if err := Convert_v1_HorizontalPodAutoscalerSpec_To_extensions_HorizontalPodAutoscalerSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_HorizontalPodAutoscalerStatus_To_extensions_HorizontalPodAutoscalerStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_HorizontalPodAutoscaler_To_extensions_HorizontalPodAutoscaler(in *HorizontalPodAutoscaler, out *extensions.HorizontalPodAutoscaler, s conversion.Scope) error {
	return autoConvert_v1_HorizontalPodAutoscaler_To_extensions_HorizontalPodAutoscaler(in, out, s)
}

func autoConvert_extensions_HorizontalPodAutoscaler_To_v1_HorizontalPodAutoscaler(in *extensions.HorizontalPodAutoscaler, out *HorizontalPodAutoscaler, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*extensions.HorizontalPodAutoscaler))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	// TODO: Inefficient conversion - can we improve it?
	if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
		return err
	}
	if err := Convert_extensions_HorizontalPodAutoscalerSpec_To_v1_HorizontalPodAutoscalerSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_extensions_HorizontalPodAutoscalerStatus_To_v1_HorizontalPodAutoscalerStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_extensions_HorizontalPodAutoscaler_To_v1_HorizontalPodAutoscaler(in *extensions.HorizontalPodAutoscaler, out *HorizontalPodAutoscaler, s conversion.Scope) error {
	return autoConvert_extensions_HorizontalPodAutoscaler_To_v1_HorizontalPodAutoscaler(in, out, s)
}

func autoConvert_v1_HorizontalPodAutoscalerList_To_extensions_HorizontalPodAutoscalerList(in *HorizontalPodAutoscalerList, out *extensions.HorizontalPodAutoscalerList, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*HorizontalPodAutoscalerList))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]extensions.HorizontalPodAutoscaler, len(*in))
		for i := range *in {
			if err := Convert_v1_HorizontalPodAutoscaler_To_extensions_HorizontalPodAutoscaler(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_v1_HorizontalPodAutoscalerList_To_extensions_HorizontalPodAutoscalerList(in *HorizontalPodAutoscalerList, out *extensions.HorizontalPodAutoscalerList, s conversion.Scope) error {
	return autoConvert_v1_HorizontalPodAutoscalerList_To_extensions_HorizontalPodAutoscalerList(in, out, s)
}

func autoConvert_extensions_HorizontalPodAutoscalerList_To_v1_HorizontalPodAutoscalerList(in *extensions.HorizontalPodAutoscalerList, out *HorizontalPodAutoscalerList, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*extensions.HorizontalPodAutoscalerList))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	if err := api.Convert_unversioned_ListMeta_To_unversioned_ListMeta(&in.ListMeta, &out.ListMeta, s); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HorizontalPodAutoscaler, len(*in))
		for i := range *in {
			if err := Convert_extensions_HorizontalPodAutoscaler_To_v1_HorizontalPodAutoscaler(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func Convert_extensions_HorizontalPodAutoscalerList_To_v1_HorizontalPodAutoscalerList(in *extensions.HorizontalPodAutoscalerList, out *HorizontalPodAutoscalerList, s conversion.Scope) error {
	return autoConvert_extensions_HorizontalPodAutoscalerList_To_v1_HorizontalPodAutoscalerList(in, out, s)
}

func autoConvert_v1_HorizontalPodAutoscalerStatus_To_extensions_HorizontalPodAutoscalerStatus(in *HorizontalPodAutoscalerStatus, out *extensions.HorizontalPodAutoscalerStatus, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*HorizontalPodAutoscalerStatus))(in)
	}
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = *in
	} else {
		out.ObservedGeneration = nil
	}
	if in.LastScaleTime != nil {
		in, out := &in.LastScaleTime, &out.LastScaleTime
		*out = new(unversioned.Time)
		if err := api.Convert_unversioned_Time_To_unversioned_Time(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.LastScaleTime = nil
	}
	out.CurrentReplicas = in.CurrentReplicas
	out.DesiredReplicas = in.DesiredReplicas
	if in.CurrentCPUUtilizationPercentage != nil {
		in, out := &in.CurrentCPUUtilizationPercentage, &out.CurrentCPUUtilizationPercentage
		*out = *in
	} else {
		out.CurrentCPUUtilizationPercentage = nil
	}
	return nil
}

func Convert_v1_HorizontalPodAutoscalerStatus_To_extensions_HorizontalPodAutoscalerStatus(in *HorizontalPodAutoscalerStatus, out *extensions.HorizontalPodAutoscalerStatus, s conversion.Scope) error {
	return autoConvert_v1_HorizontalPodAutoscalerStatus_To_extensions_HorizontalPodAutoscalerStatus(in, out, s)
}

func autoConvert_extensions_HorizontalPodAutoscalerStatus_To_v1_HorizontalPodAutoscalerStatus(in *extensions.HorizontalPodAutoscalerStatus, out *HorizontalPodAutoscalerStatus, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*extensions.HorizontalPodAutoscalerStatus))(in)
	}
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = *in
	} else {
		out.ObservedGeneration = nil
	}
	if in.LastScaleTime != nil {
		in, out := &in.LastScaleTime, &out.LastScaleTime
		*out = new(unversioned.Time)
		if err := api.Convert_unversioned_Time_To_unversioned_Time(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.LastScaleTime = nil
	}
	out.CurrentReplicas = in.CurrentReplicas
	out.DesiredReplicas = in.DesiredReplicas
	if in.CurrentCPUUtilizationPercentage != nil {
		in, out := &in.CurrentCPUUtilizationPercentage, &out.CurrentCPUUtilizationPercentage
		*out = *in
	} else {
		out.CurrentCPUUtilizationPercentage = nil
	}
	return nil
}

func Convert_extensions_HorizontalPodAutoscalerStatus_To_v1_HorizontalPodAutoscalerStatus(in *extensions.HorizontalPodAutoscalerStatus, out *HorizontalPodAutoscalerStatus, s conversion.Scope) error {
	return autoConvert_extensions_HorizontalPodAutoscalerStatus_To_v1_HorizontalPodAutoscalerStatus(in, out, s)
}

func autoConvert_v1_Scale_To_autoscaling_Scale(in *Scale, out *autoscaling.Scale, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*Scale))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	// TODO: Inefficient conversion - can we improve it?
	if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
		return err
	}
	if err := Convert_v1_ScaleSpec_To_autoscaling_ScaleSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1_ScaleStatus_To_autoscaling_ScaleStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1_Scale_To_autoscaling_Scale(in *Scale, out *autoscaling.Scale, s conversion.Scope) error {
	return autoConvert_v1_Scale_To_autoscaling_Scale(in, out, s)
}

func autoConvert_autoscaling_Scale_To_v1_Scale(in *autoscaling.Scale, out *Scale, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*autoscaling.Scale))(in)
	}
	if err := api.Convert_unversioned_TypeMeta_To_unversioned_TypeMeta(&in.TypeMeta, &out.TypeMeta, s); err != nil {
		return err
	}
	// TODO: Inefficient conversion - can we improve it?
	if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
		return err
	}
	if err := Convert_autoscaling_ScaleSpec_To_v1_ScaleSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_autoscaling_ScaleStatus_To_v1_ScaleStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

func Convert_autoscaling_Scale_To_v1_Scale(in *autoscaling.Scale, out *Scale, s conversion.Scope) error {
	return autoConvert_autoscaling_Scale_To_v1_Scale(in, out, s)
}

func autoConvert_v1_ScaleSpec_To_autoscaling_ScaleSpec(in *ScaleSpec, out *autoscaling.ScaleSpec, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*ScaleSpec))(in)
	}
	out.Replicas = in.Replicas
	return nil
}

func Convert_v1_ScaleSpec_To_autoscaling_ScaleSpec(in *ScaleSpec, out *autoscaling.ScaleSpec, s conversion.Scope) error {
	return autoConvert_v1_ScaleSpec_To_autoscaling_ScaleSpec(in, out, s)
}

func autoConvert_autoscaling_ScaleSpec_To_v1_ScaleSpec(in *autoscaling.ScaleSpec, out *ScaleSpec, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*autoscaling.ScaleSpec))(in)
	}
	out.Replicas = in.Replicas
	return nil
}

func Convert_autoscaling_ScaleSpec_To_v1_ScaleSpec(in *autoscaling.ScaleSpec, out *ScaleSpec, s conversion.Scope) error {
	return autoConvert_autoscaling_ScaleSpec_To_v1_ScaleSpec(in, out, s)
}

func autoConvert_v1_ScaleStatus_To_autoscaling_ScaleStatus(in *ScaleStatus, out *autoscaling.ScaleStatus, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*ScaleStatus))(in)
	}
	out.Replicas = in.Replicas
	out.Selector = in.Selector
	return nil
}

func Convert_v1_ScaleStatus_To_autoscaling_ScaleStatus(in *ScaleStatus, out *autoscaling.ScaleStatus, s conversion.Scope) error {
	return autoConvert_v1_ScaleStatus_To_autoscaling_ScaleStatus(in, out, s)
}

func autoConvert_autoscaling_ScaleStatus_To_v1_ScaleStatus(in *autoscaling.ScaleStatus, out *ScaleStatus, s conversion.Scope) error {
	if defaulting, found := s.DefaultingInterface(reflect.TypeOf(*in)); found {
		defaulting.(func(*autoscaling.ScaleStatus))(in)
	}
	out.Replicas = in.Replicas
	out.Selector = in.Selector
	return nil
}

func Convert_autoscaling_ScaleStatus_To_v1_ScaleStatus(in *autoscaling.ScaleStatus, out *ScaleStatus, s conversion.Scope) error {
	return autoConvert_autoscaling_ScaleStatus_To_v1_ScaleStatus(in, out, s)
}
