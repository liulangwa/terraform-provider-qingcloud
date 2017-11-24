package qingcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func modifyVxnetAttributes(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).vxnet
	input := new(qc.ModifyVxNetAttributesInput)
	input.VxNet = qc.String(d.Id())
	nameUpdate := false
	descriptionUpdate := false
	input.VxNetName, nameUpdate = getNamePointer(d)
	input.Description, descriptionUpdate = getDescriptionPointer(d)
	if nameUpdate || descriptionUpdate {
		var output *qc.ModifyVxNetAttributesOutput
		var err error
		simpleRetry(func() error {
			output, err = clt.ModifyVxNetAttributes(input)
			return isServerBusy(err)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func vxnetJoinRouter(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).router
	input := new(qc.JoinRouterInput)
	input.VxNet = qc.String(d.Id())
	input.Router = qc.String(d.Get("vpc_id").(string))
	input.IPNetwork = qc.String(d.Get("ip_network").(string))
	var output *qc.JoinRouterOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.JoinRouter(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if _, err := RouterTransitionStateRefresh(meta.(*QingCloudClient).router, d.Get("vpc_id").(string)); err != nil {
		return err
	}
	return nil
}

func vxnetLeaverRouter(d *schema.ResourceData, meta interface{}) error {
	oldVPC, _ := d.GetChange("vpc_id")
	clt := meta.(*QingCloudClient).router
	input := new(qc.LeaveRouterInput)
	input.VxNets = []*string{qc.String(d.Id())}
	input.Router = qc.String(oldVPC.(string))
	var output *qc.LeaveRouterOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.LeaveRouter(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if _, err := VxnetLeaveRouterTransitionStateRefresh(meta.(*QingCloudClient).vxnet, d.Id()); err != nil {
		return err
	}
	if _, err := RouterTransitionStateRefresh(meta.(*QingCloudClient).router, d.Get("vpc_id").(string)); err != nil {
		return err
	}
	return nil
}

func isVxnetSelfManaged(vxnetId string, clt *qc.VxNetService) (bool, error) {
	input := new(qc.DescribeVxNetsInput)
	input.VxNets = []*string{qc.String(vxnetId)}
	output, err := clt.DescribeVxNets(input)
	if err != nil {
		return false, err
	}
	if len(output.VxNetSet) == 0 {
		return false, fmt.Errorf("Error can not find vxnet ")
	}
	if qc.IntValue(output.VxNetSet[0].VxNetType) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}
