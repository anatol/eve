// Copyright (c) 2017-2022 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package zedagent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	zconfig "github.com/lf-edge/eve/api/go/config"
	"github.com/lf-edge/eve/pkg/pillar/base"
	"github.com/lf-edge/eve/pkg/pillar/flextimer"
	"github.com/lf-edge/eve/pkg/pillar/hardware"
	"github.com/lf-edge/eve/pkg/pillar/pidfile"
	"github.com/lf-edge/eve/pkg/pillar/pubsub"
	"github.com/lf-edge/eve/pkg/pillar/types"
	"github.com/lf-edge/eve/pkg/pillar/utils"
	fileutils "github.com/lf-edge/eve/pkg/pillar/utils/file"
	"github.com/lf-edge/eve/pkg/pillar/zedcloud"
	uuid "github.com/satori/go.uuid"
)

// This is set once at init time and not changed
var serverName string
var serverNameAndPort string

// Notify simple struct to pass notification messages
type Notify struct{}

// localServerAddr contains a source IP and a destination URL (without path)
// to use to connect to a particular local server.
type localServerAddr struct {
	bridgeIP        net.IP
	localServerAddr string
	appUUID         uuid.UUID
}

// localServerMap is a map of all local (profile, radio, ...) servers
type localServerMap struct {
	servers  map[string][]localServerAddr // key = bridge name, value = local servers
	upToDate bool
}

// L2Adapter is used to represent L2 Adapter (VLAN, bond) during configuration parsing.
type L2Adapter struct {
	config         *types.NetworkPortConfig
	lowerL2Ports   []*L2Adapter
	lowerPhysPorts []*types.PhysicalIOAdapter
}

type getconfigContext struct {
	zedagentCtx               *zedagentContext    // Cross link
	ledBlinkCount             types.LedBlinkCount // Current count
	configReceived            bool
	configGetStatus           types.ConfigGetStatus
	updateInprogress          bool
	readSavedConfig           bool // Did we already read it?
	configTickerHandle        interface{}
	metricsTickerHandle       interface{}
	locationCloudTickerHandle interface{}
	locationAppTickerHandle   interface{}
	localProfileTickerHandle  interface{}
	pubDevicePortConfig       pubsub.Publication
	pubPhysicalIOAdapters     pubsub.Publication
	devicePortConfig          types.DevicePortConfig
	pubNetworkXObjectConfig   pubsub.Publication
	subAppInstanceStatus      pubsub.Subscription
	subDomainMetric           pubsub.Subscription
	subProcessMetric          pubsub.Subscription
	subHostMemory             pubsub.Subscription
	subNodeAgentStatus        pubsub.Subscription
	pubZedAgentStatus         pubsub.Publication
	pubAppInstanceConfig      pubsub.Publication
	pubAppNetworkConfig       pubsub.Publication
	subAppNetworkStatus       pubsub.Subscription
	pubBaseOsConfig           pubsub.Publication
	pubBaseOs                 pubsub.Publication
	pubDatastoreConfig        pubsub.Publication
	pubNetworkInstanceConfig  pubsub.Publication
	pubControllerCert         pubsub.Publication
	pubCipherContext          pubsub.Publication
	subContentTreeStatus      pubsub.Subscription
	pubContentTreeConfig      pubsub.Publication
	subVolumeStatus           pubsub.Subscription
	pubVolumeConfig           pubsub.Publication
	pubDisksConfig            pubsub.Publication
	NodeAgentStatus           *types.NodeAgentStatus
	configProcessingSkipFlag  bool
	lastReceivedConfig        time.Time
	lastProcessedConfig       time.Time
	localProfileServer        string
	profileServerToken        string
	currentProfile            string
	globalProfile             string
	localProfile              string
	localProfileTrigger       chan Notify
	localServerMap            *localServerMap
	lastDevCmdTimestamp       uint64 // From lastDevCmdTimestampFile

	// parsed L2 adapters
	vlans []L2Adapter
	bonds []L2Adapter

	// radio-silence
	radioSilence     types.RadioSilence // the intended state of radio devices
	triggerRadioPOST chan Notify

	localAppInfoPOSTTicker flextimer.FlexTickerHandle
	localDevInfoPOSTTicker flextimer.FlexTickerHandle

	// When enabled, device location reports are being published to the Local profile server
	// at a significantly decreased rate.
	lpsThrottledLocation     bool
	lpsLastPublishedLocation time.Time

	// localCommands : list of commands requested from a local server.
	// This information is persisted under /persist/checkpoint/localcommands
	localCommands *types.LocalCommands

	callProcessLocalProfileServerChange bool //did we already call processLocalProfileServerChange

	configRetryUpdateCounter uint32 // received from config

	// Frequency in seconds at which metrics is published to the controller.
	// This value can be different from 'timer.metric.interval' in the case of
	// timer.metric.interval > currentMetricInterval, until the value of
	// 'timer.metric.interval' has been successfully notified to the controller.
	currentMetricInterval uint32

	configEdgeview *types.EdgeviewConfig // edge-view config save
}

// current devUUID from OnboardingStatus
var devUUID uuid.UUID

// Really a constant
var nilUUID uuid.UUID

// current epoch received from controller
var controllerEpoch int64

func handleConfigInit(networkSendTimeout uint32, agentMetrics *zedcloud.AgentMetrics) *zedcloud.ZedCloudContext {

	// get the server name
	bytes, err := ioutil.ReadFile(types.ServerFileName)
	if err != nil {
		log.Fatal(err)
	}
	serverNameAndPort = strings.TrimSpace(string(bytes))
	serverName = strings.Split(serverNameAndPort, ":")[0]

	zedcloudCtx := zedcloud.NewContext(log, zedcloud.ContextOptions{
		DevNetworkStatus: deviceNetworkStatus,
		Timeout:          networkSendTimeout,
		AgentMetrics:     agentMetrics,
		Serial:           hardware.GetProductSerial(log),
		SoftSerial:       hardware.GetSoftSerial(log),
		AgentName:        agentName,
	})

	log.Functionf("Configure Get Device Serial %s, Soft Serial %s, Use V2 API %v", zedcloudCtx.DevSerial,
		zedcloudCtx.DevSoftSerial, zedcloud.UseV2API())

	// XXX need to redo this since the root certificates can change
	err = zedcloud.UpdateTLSConfig(&zedcloudCtx, serverName, nil)
	if err != nil {
		log.Fatal(err)
	}

	zedcloudCtx.DevUUID = devUUID
	return &zedcloudCtx
}

// Run a periodic fetch of the config
func configTimerTask(handleChannel chan interface{},
	getconfigCtx *getconfigContext) {

	ctx := getconfigCtx.zedagentCtx
	configUrl := zedcloud.URLPathString(serverNameAndPort, zedcloudCtx.V2API, devUUID, "config")
	iteration := 0
	configProcessingSkipFlag := getLatestConfig(configUrl, iteration,
		getconfigCtx)
	if configProcessingSkipFlag != getconfigCtx.configProcessingSkipFlag {
		getconfigCtx.configProcessingSkipFlag = configProcessingSkipFlag
		triggerPublishDevInfo(ctx)
	}
	getconfigCtx.localServerMap.upToDate = false
	publishZedAgentStatus(getconfigCtx)

	configInterval := ctx.globalConfig.GlobalValueInt(types.ConfigInterval)
	interval := time.Duration(configInterval) * time.Second
	max := float64(interval)
	min := max * 0.3
	ticker := flextimer.NewRangeTicker(time.Duration(min),
		time.Duration(max))
	// Return handle to caller
	handleChannel <- ticker

	// ticker for periodical info publish around 10 min when no real change
	interval2 := time.Duration(600) * time.Second
	max2 := float64(interval2) * 1.2
	min2 := float64(interval2) * 0.8
	tickerInfo := flextimer.NewRangeTicker(time.Duration(min2),
		time.Duration(max2))

	wdName := agentName + "config"

	// Run a periodic timer so we always update StillRunning
	stillRunning := time.NewTicker(25 * time.Second)
	ctx.ps.StillRunning(wdName, warningTime, errorTime)
	ctx.ps.RegisterFileWatchdog(wdName)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			iteration += 1
			// In case devUUID changed we re-generate
			configUrl = zedcloud.URLPathString(serverNameAndPort,
				zedcloudCtx.V2API, devUUID, "config")
			configProcessingSkipFlag := getLatestConfig(configUrl, iteration, getconfigCtx)
			if configProcessingSkipFlag != getconfigCtx.configProcessingSkipFlag {
				getconfigCtx.configProcessingSkipFlag = configProcessingSkipFlag
				triggerPublishDevInfo(ctx)
			}
			getconfigCtx.localServerMap.upToDate = false
			ctx.ps.CheckMaxTimeTopic(wdName, "getLastestConfig", start,
				warningTime, errorTime)
			publishZedAgentStatus(getconfigCtx)

		case <-tickerInfo.C:
			start := time.Now()
			triggerPublishDevInfo(ctx)
			ctx.ps.CheckMaxTimeTopic(wdName, "publishInfoTimer", start,
				warningTime, errorTime)

		case <-stillRunning.C:
			if getconfigCtx.configProcessingSkipFlag {
				log.Noticef("config processing skip flag set")
			}
		}
		ctx.ps.StillRunning(wdName, warningTime, errorTime)
	}
}

func triggerGetConfig(tickerHandle interface{}) {
	log.Functionf("triggerGetConfig()")
	flextimer.TickNow(tickerHandle)
}

// Called when globalConfig changes
// Assumes the caller has verifier that the interval has changed
func updateConfigTimer(configInterval uint32, tickerHandle interface{}) {

	if tickerHandle == nil {
		// Happens if we have a GlobalConfig setting in /persist/
		log.Warnf("updateConfigTimer: no configTickerHandle yet")
		return
	}
	interval := time.Duration(configInterval) * time.Second
	log.Functionf("updateConfigTimer() change to %v", interval)
	max := float64(interval)
	min := max * 0.3
	flextimer.UpdateRangeTicker(tickerHandle,
		time.Duration(min), time.Duration(max))
	// Force an immediate timout since timer could have decreased
	flextimer.TickNow(tickerHandle)
}

// Start by trying the all the free management ports and then all the non-free
// until one succeeds in communicating with the cloud.
// We use the iteration argument to start at a different point each time.
// Returns a configProcessingSkipFlag
func getLatestConfig(url string, iteration int,
	getconfigCtx *getconfigContext) bool {

	log.Tracef("getLatestConfig(%s, %d)", url, iteration)
	// If we haven't yet published our certificates we defer to ensure
	// that the controller has our certs and can add encrypted secrets to
	// our config.
	if !getconfigCtx.zedagentCtx.publishedEdgeNodeCerts {
		log.Noticef("Defer fetching config until our EdgeNodeCerts have been published")
		return false
	}
	ctx := getconfigCtx.zedagentCtx
	const bailOnHTTPErr = false // For 4xx and 5xx HTTP errors we try other interfaces
	// except http.StatusForbidden(which returns error
	// irrespective of bailOnHTTPErr)
	getconfigCtx.configGetStatus = types.ConfigGetFail
	b, cr, err := generateConfigRequest(getconfigCtx)
	if err != nil {
		// XXX	fatal?
		return false
	}
	buf := bytes.NewBuffer(b)
	size := int64(proto.Size(cr))
	ctxWork, cancel := zedcloud.GetContextForAllIntfFunctions(zedcloudCtx)
	defer cancel()
	resp, contents, senderStatus, err := zedcloud.SendOnAllIntf(ctxWork, zedcloudCtx, url, size, buf, iteration, bailOnHTTPErr)
	if err != nil {
		newCount := types.LedBlinkConnectingToController
		switch senderStatus {
		case types.SenderStatusUpgrade:
			log.Functionf("getLatestConfig : Controller upgrade in progress")
		case types.SenderStatusRefused:
			log.Functionf("getLatestConfig : Controller returned ECONNREFUSED")
		case types.SenderStatusCertInvalid:
			log.Warnf("getLatestConfig : Controller certificate invalid time")
		case types.SenderStatusCertMiss:
			log.Functionf("getLatestConfig : Controller certificate miss")
		case types.SenderStatusNotFound:
			log.Functionf("getLatestConfig : Device deleted in controller?")
		case types.SenderStatusForbidden:
			log.Functionf("getLatestConfig : Device integrity token mismatch")
		default:
			log.Errorf("getLatestConfig  failed: %s", err)
		}
		switch senderStatus {
		case types.SenderStatusUpgrade, types.SenderStatusRefused, types.SenderStatusCertInvalid, types.SenderStatusNotFound:
			newCount = types.LedBlinkConnectedToController // Almost connected to controller!
			// Don't treat as upgrade failure
			if getconfigCtx.updateInprogress {
				log.Warnf("remoteTemporaryFailure don't fail update")
				getconfigCtx.configGetStatus = types.ConfigGetTemporaryFail
			}
		case types.SenderStatusCertMiss:
			// trigger to acquire new controller certs from cloud
			triggerControllerCertEvent(ctx)
		}
		if getconfigCtx.ledBlinkCount == types.LedBlinkOnboarded {
			// Inform ledmanager about loss of config from cloud
			utils.UpdateLedManagerConfig(log, newCount)
			getconfigCtx.ledBlinkCount = newCount
		}
		if senderStatus == types.SenderStatusNotFound {
			potentialUUIDUpdate(getconfigCtx)
		}
		if senderStatus == types.SenderStatusForbidden {
			log.Errorf("Config request is forbidden, triggering attestation again")
			_ = restartAttestation(ctx)
			if getconfigCtx.updateInprogress {
				log.Warnf("updateInprogress=true,resp.StatusCode=Forbidden, so marking ConfigGetTemporaryFail")
				getconfigCtx.configGetStatus = types.ConfigGetTemporaryFail
			}
		}

		if !getconfigCtx.readSavedConfig && !getconfigCtx.configReceived {
			// If we didn't yet get a config, then look for a file
			// XXX should we try a few times?
			// If we crashed we wait until we connect to zedcloud so that
			// keyboard can be enabled and things can be debugged and not
			// have e.g., an OOM reboot loop
			if !ctx.bootReason.StartWithSavedConfig() {
				log.Warnf("Ignore any saved config due to boot reason %s",
					ctx.bootReason)
			} else {
				config, ts, err := readSavedProtoMessageConfig(
					ctx.globalConfig.GlobalValueInt(types.StaleConfigTime),
					checkpointDirname+"/lastconfig", false)
				if err != nil {
					log.Errorf("getconfig: %v", err)
					return false
				}
				if config != nil {
					log.Noticef("Using saved config dated %s",
						ts.Format(time.RFC3339Nano))
					getconfigCtx.readSavedConfig = true
					getconfigCtx.configGetStatus = types.ConfigGetReadSaved
					return inhaleDeviceConfig(config, getconfigCtx,
						true)
				}
			}
		}
		publishZedAgentStatus(getconfigCtx)
		return false
	}

	if resp.StatusCode == http.StatusNotModified {
		log.Tracef("StatusNotModified len %d", len(contents))
		// Inform ledmanager about config received from cloud
		utils.UpdateLedManagerConfig(log, types.LedBlinkOnboarded)
		getconfigCtx.ledBlinkCount = types.LedBlinkOnboarded

		if !getconfigCtx.configReceived {
			getconfigCtx.configReceived = true
		}
		getconfigCtx.configGetStatus = types.ConfigGetSuccess
		publishZedAgentStatus(getconfigCtx)

		log.Tracef("Configuration from zedcloud is unchanged")
		// Update modification time since checked by readSavedConfig
		touchReceivedProtoMessage()
		return false
	}

	if err := zedcloud.ValidateProtoContentType(url, resp); err != nil {
		log.Errorln("validateProtoMessage: ", err)
		// Inform ledmanager about cloud connectivity
		utils.UpdateLedManagerConfig(log, types.LedBlinkConnectedToController)
		getconfigCtx.ledBlinkCount = types.LedBlinkConnectedToController
		publishZedAgentStatus(getconfigCtx)
		return false
	}

	changed, config, err := readConfigResponseProtoMessage(resp, contents)
	if err != nil {
		log.Errorln("readConfigResponseProtoMessage: ", err)
		// Inform ledmanager about cloud connectivity
		utils.UpdateLedManagerConfig(log, types.LedBlinkConnectedToController)
		getconfigCtx.ledBlinkCount = types.LedBlinkConnectedToController
		publishZedAgentStatus(getconfigCtx)
		return false
	}

	// Inform ledmanager about config received from cloud
	utils.UpdateLedManagerConfig(log, types.LedBlinkOnboarded)
	getconfigCtx.ledBlinkCount = types.LedBlinkOnboarded

	if !getconfigCtx.configReceived {
		getconfigCtx.configReceived = true
	}
	getconfigCtx.configGetStatus = types.ConfigGetSuccess
	publishZedAgentStatus(getconfigCtx)

	if !changed {
		log.Tracef("Configuration from zedcloud is unchanged")
		// Update modification time since checked by readSavedConfig
		touchReceivedProtoMessage()
		return false
	}
	saveReceivedProtoMessage(contents)

	return inhaleDeviceConfig(config, getconfigCtx, false)
}

func saveReceivedProtoMessage(contents []byte) {
	saveConfig("lastconfig", contents)
}

// Update timestamp - no content changes
func touchReceivedProtoMessage() {
	touchSavedConfig("lastconfig")
}

// XXX for debug we track these
func saveSentMetricsProtoMessage(contents []byte) {
	saveConfig("lastmetrics", contents)
}

// XXX for debug we track these
func saveSentDeviceInfoProtoMessage(contents []byte) {
	saveConfig("lastdeviceinfo", contents)
}

// XXX for debug we track these
func saveSentAppInfoProtoMessage(contents []byte) {
	saveConfig("lastappinfo", contents)
}

func saveConfig(filename string, contents []byte) {
	filename = checkpointDirname + "/" + filename
	err := fileutils.WriteRename(filename, contents)
	if err != nil {
		// Can occur if no space in filesystem
		log.Errorf("saveConfig failed: %s", err)
		return
	}
}

// Remove saved config file if it exists.
func cleanSavedConfig(filename string) {
	filename = checkpointDirname + "/" + filename
	if err := os.Remove(filename); err != nil {
		log.Functionf("cleanSavedConfig failed: %s", err)
	}
}

// Update modification time
func touchSavedConfig(filename string) {
	filename = checkpointDirname + "/" + filename
	_, err := os.Stat(filename)
	if err != nil {
		log.Warnf("touchSavedConfig stat failed: %s", err)
	}
	currentTime := time.Now()
	err = os.Chtimes(filename, currentTime, currentTime)
	if err != nil {
		// Can occur if no space in filesystem?
		log.Errorf("touchSavedConfig failed: %s", err)
	}
}

// If the file exists then read the config, and return is modify time
// Ignore if older than StaleConfigTime seconds
func readSavedProtoMessageConfig(staleConfigTime uint32,
	filename string, force bool) (*zconfig.EdgeDevConfig, time.Time, error) {
	contents, ts, err := readSavedConfig(staleConfigTime, filename, force)
	if err != nil {
		log.Errorln("readSavedProtoMessageConfig", err)
		return nil, ts, err
	}
	var configResponse = &zconfig.ConfigResponse{}
	err = proto.Unmarshal(contents, configResponse)
	if err != nil {
		log.Errorf("readSavedProtoMessageConfig Unmarshalling failed: %v",
			err)
		return nil, ts, err
	}
	config := configResponse.GetConfig()
	return config, ts, nil
}

// If the file exists then read the config content from it, and return its modify time.
// Ignore if older than staleTime seconds.
func readSavedConfig(staleTime uint32,
	filename string, force bool) ([]byte, time.Time, error) {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) && !force {
			return nil, time.Time{}, nil
		} else {
			return nil, time.Time{}, err
		}
	}
	age := time.Since(info.ModTime())
	staleLimit := time.Second * time.Duration(staleTime)
	if !force && age > staleLimit {
		errStr := fmt.Sprintf("saved config too old: age %v limit %d\n",
			age, staleLimit)
		log.Errorln(errStr)
		return nil, info.ModTime(), nil
	}
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorln("readSavedConfig", err)
		return nil, info.ModTime(), err
	}
	return contents, info.ModTime(), nil
}

// The most recent config hash we received. Starts empty
var prevConfigHash string

func generateConfigRequest(getconfigCtx *getconfigContext) ([]byte, *zconfig.ConfigRequest, error) {
	log.Tracef("generateConfigRequest() sending hash %s", prevConfigHash)
	configRequest := &zconfig.ConfigRequest{
		ConfigHash: prevConfigHash,
	}
	//Populate integrity token if there is one available
	iToken, err := readIntegrityToken()
	if err == nil {
		configRequest.IntegrityToken = iToken
	}
	b, err := proto.Marshal(configRequest)
	if err != nil {
		log.Errorln(err)
		return nil, nil, err
	}
	return b, configRequest, nil
}

// Returns changed, config, error. The changed is based the ConfigRequest vs
// the ConfigResponse hash
func readConfigResponseProtoMessage(resp *http.Response, contents []byte) (bool, *zconfig.EdgeDevConfig, error) {

	var configResponse = &zconfig.ConfigResponse{}
	err := proto.Unmarshal(contents, configResponse)
	if err != nil {
		log.Errorf("Unmarshalling failed: %v", err)
		return false, nil, err
	}
	hash := configResponse.GetConfigHash()
	if hash == prevConfigHash {
		log.Tracef("Same ConfigHash %s len %d", hash, len(contents))
		return false, nil, nil
	}
	log.Tracef("Change in ConfigHash from %s to %s", prevConfigHash, hash)
	prevConfigHash = hash
	config := configResponse.GetConfig()
	return true, config, nil
}

// Returns a configProcessingSkipFlag
func inhaleDeviceConfig(config *zconfig.EdgeDevConfig, getconfigCtx *getconfigContext, usingSaved bool) bool {
	log.Tracef("Inhaling config")

	// if they match return
	var devId = &zconfig.UUIDandVersion{}

	devId = config.GetId()
	if devId != nil {
		id, err := uuid.FromString(devId.Uuid)
		if err != nil {
			log.Errorf("Invalid UUID %s from cloud: %s",
				devId.Uuid, err)
			return false
		}
		if id != devUUID {
			log.Warnf("Device UUID changed from %s to %s",
				devUUID.String(), id.String())
			potentialUUIDUpdate(getconfigCtx)
			return false
		}
		newControllerEpoch := config.GetControllerEpoch()
		if controllerEpoch != newControllerEpoch {
			log.Noticef("Controller epoch changed from %d to %d", controllerEpoch, newControllerEpoch)
			controllerEpoch = newControllerEpoch
			triggerPublishAllInfo(getconfigCtx.zedagentCtx)
		}
	}

	// add new BaseOS/App instances; returns configProcessingSkipFlag
	return parseConfig(config, getconfigCtx, usingSaved)
}

var (
	lastDevUUIDChange       = time.Now()
	potentialUUIDUpdateLock sync.Mutex
)

// When we think (due to 404) or know that the controller has changed our UUID,
// ask client to get it so OnboardingStatus can be updated and notified to all agents
// The controller might do this due to a delete and re-onboard with the same device
// certificate.
// We ask client at most every 10 minutes.
// We check that another zedclient instance is not running
func potentialUUIDUpdate(_ *getconfigContext) {
	potentialUUIDUpdateLock.Lock()
	if time.Since(lastDevUUIDChange) < 10*time.Minute {
		log.Warnf("Device UUID last changed %v ago",
			time.Since(lastDevUUIDChange))
		potentialUUIDUpdateLock.Unlock()
		return
	}
	if exists, description := pidfile.CheckProcessExists(log, "zedclient"); exists {
		log.Warnf("another process is still running: %s", description)
		potentialUUIDUpdateLock.Unlock()
		return
	}
	lastDevUUIDChange = time.Now()
	// after time updated we can unlock mutex to go into time check from other routine
	potentialUUIDUpdateLock.Unlock()
	cmd := "/opt/zededa/bin/client"
	cmdArgs := []string{"getUuid"}
	log.Noticef("Calling command %s %v", cmd, cmdArgs)
	out, err := base.Exec(log, cmd, cmdArgs...).CombinedOutput()
	if err != nil {
		log.Errorf("client command %s failed %s output %s",
			cmdArgs, err, out)
	}
}

func publishZedAgentStatus(getconfigCtx *getconfigContext) {
	ctx := getconfigCtx.zedagentCtx
	status := types.ZedAgentStatus{
		Name:                 agentName,
		ConfigGetStatus:      getconfigCtx.configGetStatus,
		RebootCmd:            ctx.rebootCmd,
		ShutdownCmd:          ctx.shutdownCmd,
		PoweroffCmd:          ctx.poweroffCmd,
		RebootReason:         ctx.currentRebootReason,
		BootReason:           ctx.currentBootReason,
		MaintenanceMode:      ctx.maintenanceMode,
		ForceFallbackCounter: ctx.forceFallbackCounter,
		CurrentProfile:       getconfigCtx.currentProfile,
		RadioSilence:         getconfigCtx.radioSilence,
	}
	pub := getconfigCtx.pubZedAgentStatus
	pub.Publish(agentName, status)
}

// updateLocalServerMap processes configuration of network instances to locate all local servers matching
// the given localServerURL.
// Returns the source IP and a normalized URL for one or more network instances on which the local server
// was found to be hosted.
func updateLocalServerMap(getconfigCtx *getconfigContext, localServerURL string) error {
	url, err := url.Parse(localServerURL)
	if err != nil {
		return fmt.Errorf("updateLocalServerMap: url.Parse: %v", err)
	}

	srvMap := &localServerMap{servers: make(map[string][]localServerAddr), upToDate: true}
	appNetworkStatuses := getconfigCtx.subAppNetworkStatus.GetAll()
	networkInstanceConfigs := getconfigCtx.pubNetworkInstanceConfig.GetAll()
	localServerHostname := url.Hostname()
	localServerIP := net.ParseIP(localServerHostname)

	for _, entry := range appNetworkStatuses {
		appNetworkStatus := entry.(types.AppNetworkStatus)
		for _, ulStatus := range appNetworkStatus.UnderlayNetworkList {
			bridgeIP := net.ParseIP(ulStatus.BridgeIPAddr)
			if bridgeIP == nil {
				continue
			}
			if localServerIP != nil {
				// check if the defined IP of localServer equals the allocated IP of the app
				if ulStatus.AllocatedIPv4Addr == localServerIP.String() {
					srvAddr := localServerAddr{
						localServerAddr: localServerURL,
						bridgeIP:        bridgeIP,
						appUUID:         appNetworkStatus.UUIDandVersion.UUID,
					}
					srvMap.servers[ulStatus.Bridge] = append(srvMap.servers[ulStatus.Bridge], srvAddr)
				}
				continue
			}
			// check if defined hostname of localServer is in DNS records
			for _, ni := range networkInstanceConfigs {
				networkInstanceConfig := ni.(types.NetworkInstanceConfig)
				for _, dnsNameToIPList := range networkInstanceConfig.DnsNameToIPList {
					if dnsNameToIPList.HostName != localServerHostname {
						continue
					}
					for _, ip := range dnsNameToIPList.IPs {
						localServerURLReplaced := strings.Replace(
							localServerURL, localServerHostname, ip.String(), 1)
						log.Functionf(
							"updateLocalServerMap: will use %s for bridge %s",
							localServerURLReplaced, ulStatus.Bridge)
						srvAddr := localServerAddr{
							localServerAddr: localServerURLReplaced,
							bridgeIP:        bridgeIP,
							appUUID:         appNetworkStatus.UUIDandVersion.UUID,
						}
						srvMap.servers[ulStatus.Bridge] = append(srvMap.servers[ulStatus.Bridge], srvAddr)
					}
				}
			}
		}
	}
	// To handle concurrent access to localServerMap (from localProfileTimerTask, radioPOSTTask and potentially from
	// some more future tasks), we replace the map pointer at the very end of this function once the map is fully
	// constructed.
	getconfigCtx.localServerMap = srvMap
	return nil
}

// updateHasLocalServer sets HasLocalServer on the app instances
// Note that if there are changes to the AppInstanceConfig or the allocated IP
// addresses the HasLocalServer will not immediately reflect that since we need
// the IP address from AppNetworkStatus.
func updateHasLocalServer(ctx *getconfigContext) {
	srvMap := ctx.localServerMap.servers
	items := ctx.pubAppInstanceConfig.GetAll()
	for _, item := range items {
		aic := item.(types.AppInstanceConfig)
		hasLocalServer := false
		for _, servers := range srvMap {
			for _, srv := range servers {
				if srv.appUUID == aic.UUIDandVersion.UUID {
					hasLocalServer = true
					break
				}
			}
		}
		if hasLocalServer != aic.HasLocalServer {
			aic.HasLocalServer = hasLocalServer
			log.Noticef("HasLocalServer(%s) for %s change to %t",
				aic.Key(), aic.DisplayName, hasLocalServer)
			// Verify that it fits and if not publish with error
			checkAndPublishAppInstanceConfig(ctx, aic)
		}
	}
}
