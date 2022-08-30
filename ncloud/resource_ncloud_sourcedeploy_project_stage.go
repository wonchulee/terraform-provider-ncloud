package ncloud

import (
	"fmt"
	"context"
	"regexp"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vsourcedeploy"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func init() {
	RegisterResource("ncloud_sourcedeploy_project_stage", resourceNcloudSourceDeployStage())
}

func resourceNcloudSourceDeployStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNcloudSourceDeployStageCreate,
		ReadContext:   resourceNcloudSourceDeployStageRead,
		DeleteContext: resourceNcloudSourceDeployStageDelete,
		UpdateContext: resourceNcloudSourceDeployStageUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(DefaultTimeout),
			Read:   schema.DefaultTimeout(DefaultTimeout),
			Update: schema.DefaultTimeout(DefaultTimeout),
			Delete: schema.DefaultTimeout(DefaultTimeout),
		},
		Schema: map[string]*schema.Schema{
			"project_id":{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: ToDiagFunc(validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^[^ !@#$%^&*()+\=\[\]{};':"\\|,.<>\/?]+$`), `Cannot contain special characters ( !@#$%^&*()+\=\[\]{};':"\\|,.<>\/?).`),
				)),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: ToDiagFunc(validation.StringInSlice([]string{"Server", "AutoScalingGroup", "KubernetesService", "ObjectStorage"}, false)),
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server_no": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"server_name": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"auto_scaling_group_no": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"auto_scaling_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_uuid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cluster_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"bucket_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceNcloudSourceDeployStageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	
	if !config.SupportVPC {
		return diag.FromErr(NotSupportClassic("resource `ncloud_sourcedeploy_project_stage`"))
	}

	reqParams, paramsErr := getStage(d)
	if paramsErr != nil{
		return diag.FromErr(paramsErr)
	}
	projectId := ncloud.IntString(d.Get("project_id").(int))
	logCommonRequest("createSourceDeployStage", reqParams)
	resp, err := config.Client.vsourcedeploy.V1Api.CreateStage(ctx, reqParams, projectId)
	if err != nil {
		logErrorResponse("createSourceDeployStage", err, reqParams)
		return diag.FromErr(err)
	}
	logResponse("createSourceDeployStage", resp.Id)

	d.SetId(*ncloud.IntString(int(ncloud.Int32Value(resp.Id))))
	d.Set("project_id", Int32PtrOrNil(d.GetOk("project_id")))

	return resourceNcloudSourceDeployStageRead(ctx, d, meta)
}


func resourceNcloudSourceDeployStageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)

	if !config.SupportVPC {
		return diag.FromErr(NotSupportClassic("resource `ncloud_sourcedeploy_project_stage`"))
	}
	projectId := ncloud.IntString(d.Get("project_id").(int))
	stage, err := getSourceDeployStageById(ctx, config, projectId, ncloud.String(d.Id()))

	if err != nil {
		return diag.FromErr(err)
	}

	if stage == nil {
		d.SetId("")
		return nil
	}
	
	d.SetId(*ncloud.IntString(int(ncloud.Int32Value(stage.Id))))
	d.Set("name", stage.Name)
	d.Set("type", stage.Type_)
	d.Set("config",makeStageConfig(stage.Config))

	return nil
}

func resourceNcloudSourceDeployStageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)

	err := changeDeployStage(ctx, d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNcloudSourceDeployStageRead(ctx, d, meta)
}

func resourceNcloudSourceDeployStageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	if !config.SupportVPC {
		return diag.FromErr(NotSupportClassic("resource `ncloud_sourcedeploy_project_stage`"))
	}

	projectId := ncloud.IntString(d.Get("project_id").(int))
	logCommonRequest("deleteSourceDeployStage", d.Id())
	resp, err := config.Client.vsourcedeploy.V1Api.DeleteStage(ctx, projectId, ncloud.String(d.Id()))
	if err != nil {
		logErrorResponse("deleteSourceDeployStage", err, d.Id())
		return diag.FromErr(err)
	}

	logResponse("deleteSourceDeployStage", resp)
	d.SetId("")
	return nil
}


func getStage(d *schema.ResourceData) (*vsourcedeploy.CreateStage, error){
	deployTarget,  deployTargetErr := getDeployTarget(d)
	if deployTargetErr != nil {
		return nil, deployTargetErr
	}
	reqParams := &vsourcedeploy.CreateStage{
		Name:              	StringPtrOrNil(d.GetOk("name")),
		Type_:				StringPtrOrNil(d.GetOk("type")),
		Config:				deployTarget,
	}
	return reqParams, nil
}

func getDeployTarget(d *schema.ResourceData) (*vsourcedeploy.StageConfig, error){
	deployTarget :=	vsourcedeploy.StageConfig{}

	deployTargetType := ncloud.StringValue(StringPtrOrNil(d.GetOk("type")))

	switch deployTargetType {
	case "Server":
		if serverNo, ok := d.GetOk("config.0.server_no"); ok {
			deployTarget.ServerNo = expandStringInterfaceListToInt32List(serverNo.([]interface{}))
		}
		if deployTarget.ServerNo == nil {
			return nil, fmt.Errorf("config(server_no) is required")
		}
	case "AutoScalingGroup":
		deployTarget.AutoScalingGroupNo = Int32PtrOrNil(d.GetOk("config.0.auto_scaling_group_no"))
		if deployTarget.AutoScalingGroupNo == nil{
			return nil, fmt.Errorf("config(auto_scaling_group_no) is required")
		}
	case "KubernetesService":
		deployTarget.ClusterUuid = StringPtrOrNil(d.GetOk("config.0.cluster_uuid"))
		if deployTarget.ClusterUuid == nil{
			return nil, fmt.Errorf("config(cluster_uuid) is required")
		}
	case "ObjectStorage":
		deployTarget.BucketName = StringPtrOrNil(d.GetOk("config.0.bucket_name"))
		if deployTarget.BucketName == nil{
			return nil, fmt.Errorf("config(bucket_name) is required")
		}
	}
	return &deployTarget, nil
}

func getSourceDeployStageById(ctx context.Context,config *ProviderConfig, projectId *string, id *string) (*vsourcedeploy.GetStageDetailResponse, error) {
	logCommonRequest("getSourceDeployStage", id)
	resp, err := config.Client.vsourcedeploy.V1Api.GetStage(ctx, projectId, id)
	if err != nil {
		logErrorResponse("getSourceDeployStage", err, *id)
		return nil, err
	}
	logResponse("getSourceDeployStage", resp)

	return resp, nil
}

func makeStageConfig(config *vsourcedeploy.GetStageDetailResponseConfig) []interface{}{
	if config == nil{
		return nil
	}
	values := map[string]interface{}{}

	values["server_no"] = flattenInt32ListToStringList(config.ServerNo)
	values["server_name"] = ncloud.StringListValue(config.ServerName)
	values["auto_scaling_group_no"] = ncloud.Int32Value(config.AutoScalingGroupNo)
	values["auto_scaling_group_name"] = ncloud.StringValue(config.AutoScalingGroupName)
	values["cluster_uuid"] = ncloud.StringValue(config.ClusterUuid)
	values["cluster_name"] = ncloud.StringValue(config.ClusterName)
	values["bucket_name"] = ncloud.StringValue(config.BucketName)

	return []interface{}{values}
}


func changeDeployStage(ctx context.Context, d *schema.ResourceData, config *ProviderConfig) error {

	reqParams, paramErr := getStage(d)
	
	if paramErr != nil {
		return paramErr
	}
	projectId := ncloud.IntString(d.Get("project_id").(int))
	id := ncloud.String(d.Id())

	logCommonRequest("changeSourceDeployStage", reqParams)
	resp, err := config.Client.vsourcedeploy.V1Api.ChangeStage(ctx, reqParams, projectId, id)
	if err != nil {
		logErrorResponse("changeSourceDeployStage", err, reqParams)
		return err
	}
	logResponse("changeSourceDeployStage", resp)

	return nil
}