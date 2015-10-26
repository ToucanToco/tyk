package main

import (
	"github.com/lonelycode/go-uuid/uuid"
	"github.com/lonelycode/tykcommon"
	"net/url"
	"time"
)

var GlobalHostChecker HostCheckerManager

type HostCheckerManager struct {
	Id                string
	store             StorageHandler
	checker           *HostUptimeChecker
	stopLoop          bool
	pollerStarted     bool
	unhealthyHostList map[string]bool
}

const (
	UnHealthyHostMetaDataTargetKey string = "target_url"
	UnHealthyHostMetaDataAPIKey    string = "api_id"
	UnHealthyHostMetaDataHostKey   string = "host_name"
	PollerCacheKey                 string = "PollerActiveInstanceID"
	PoolerHostSentinelKeyPrefix    string = "PollerCheckerInstance:"
)

func (hc *HostCheckerManager) Init(store StorageHandler) {
	hc.store = store
	hc.unhealthyHostList = make(map[string]bool)
	// Generate a new ID for ourselves
	hc.GenerateCheckerId()
}

func (hc *HostCheckerManager) Start() {
	// Start loop to check if we are active instance
	if hc.Id != "" {
		go hc.CheckActivePollerLoop()
	}
}

func (hc *HostCheckerManager) GenerateCheckerId() {
	hc.Id = uuid.NewUUID().String()
}

func (hc *HostCheckerManager) CheckActivePollerLoop() {
	for {
		if hc.stopLoop {
			log.Info("[HOST CHECK MANAGER] Stopping uptime tests")
			break
		}

		// If I'm polling, lets start the loop
		if hc.AmIPolling() {
			if !hc.pollerStarted {
				hc.pollerStarted = true
				go hc.StartPoller()
			}
		} else {
			log.Info("[HOST CHECK MANAGER] New master found, stopping uptime tests")
			go hc.StopPoller()
			hc.pollerStarted = false
		}

		time.Sleep(10 * time.Second)
	}
}

func (hc *HostCheckerManager) AmIPolling() bool {
	if hc.store == nil {
		log.Error("[HOST CHECK MANAGER] No storage instance set for uptime tests! Disabling poller...")
		return false
	}
	ActiveInstance, err := hc.store.GetKey(PollerCacheKey)
	if err != nil {
		log.Debug("[HOST CHECK MANAGER] No Primary instance found, assuming control")
		hc.store.SetKey(PollerCacheKey, hc.Id, 15)
		return true
	}

	if ActiveInstance == hc.Id {
		log.Debug("[HOST CHECK MANAGER] Primary instance set, I am master")
		hc.store.SetKey(PollerCacheKey, hc.Id, 15) // Reset TTL
		return true
	}

	return false
}

func (hc *HostCheckerManager) StartPoller() {

	log.Debug("---> Initialising checker")
	hostList := map[string]HostData{}

	// If we are restarting, we want to retain the host list
	if hc.checker == nil {
		hc.checker = &HostUptimeChecker{}
	} else {
		hostList = hc.checker.HostList
	}

	hc.checker.Init(config.UptimeTests.Config.CheckerPoolSize,
		config.UptimeTests.Config.FailureTriggerSampleSize,
		config.UptimeTests.Config.TimeWait,
		hostList,
		hc.OnHostDown,   // On failure
		hc.OnHostBackUp) // On success

	// Start the check loop
	log.Debug("---> Starting checker")
	hc.checker.Start()
	log.Debug("---> Checker started.")
}

func (hc *HostCheckerManager) StopPoller() {
	if hc.checker != nil {
		hc.checker.Stop()
	}
}

func (hc *HostCheckerManager) getHostKey(report HostHealthReport) string {
	return PoolerHostSentinelKeyPrefix + report.MetaData[UnHealthyHostMetaDataHostKey]
}

func (hc *HostCheckerManager) OnHostDown(report HostHealthReport) {
	log.Debug("Update key: ", hc.getHostKey(report))
	hc.store.SetKey(hc.getHostKey(report), "1", int64(config.UptimeTests.Config.TimeWait))

	thisSpec, found := ApiSpecRegister[report.MetaData[UnHealthyHostMetaDataAPIKey]]
	if !found {
		log.Warning("[HOST CHECKER MANAGER] Event can't fire for API that doesn't exist")
		return
	}

	go thisSpec.FireEvent(EVENT_HOSTDOWN,
		EVENT_HostStatusMeta{
			EventMetaDefault: EventMetaDefault{Message: "Uptime test failed"},
			HostInfo:         report,
		})

	log.Warning("[HOST CHECKER MANAGER] Host is down: ", report.CheckURL)
}

func (hc *HostCheckerManager) OnHostBackUp(report HostHealthReport) {
	log.Debug("Delete key: ", hc.getHostKey(report))
	hc.store.DeleteKey(hc.getHostKey(report))

	thisSpec, found := ApiSpecRegister[report.MetaData[UnHealthyHostMetaDataAPIKey]]
	if !found {
		log.Warning("[HOST CHECKER MANAGER] Event can't fire for API that doesn't exist")
		return
	}
	go thisSpec.FireEvent(EVENT_HOSTUP,
		EVENT_HostStatusMeta{
			EventMetaDefault: EventMetaDefault{Message: "Uptime test suceeded"},
			HostInfo:         report,
		})

	log.Warning("[HOST CHECKER MANAGER] Host is back up: ", report.CheckURL)
}

func (hc *HostCheckerManager) IsHostDown(thisUrl string) bool {
	u, err := url.Parse(thisUrl)
	if err != nil {
		log.Error(err)
	}

	log.Debug("Key is: ", PoolerHostSentinelKeyPrefix+u.Host)
	_, fErr := hc.store.GetKey(PoolerHostSentinelKeyPrefix + u.Host)

	if fErr != nil {
		return false
	}

	return true
}

func (hc *HostCheckerManager) PrepareTrackingHost(checkObject tykcommon.HostCheckObject, APIID string) HostData {
	// Build the check URL:
	u, err := url.Parse(checkObject.CheckURL)
	if err != nil {
		log.Error(err)
	}

	thisHostData := HostData{
		CheckURL: checkObject.CheckURL,
		ID:       checkObject.CheckURL,
		MetaData: make(map[string]string),
	}

	// Add our specific metadata
	thisHostData.MetaData[UnHealthyHostMetaDataTargetKey] = checkObject.CheckURL
	thisHostData.MetaData[UnHealthyHostMetaDataAPIKey] = APIID
	thisHostData.MetaData[UnHealthyHostMetaDataHostKey] = u.Host

	return thisHostData
}

func (hc *HostCheckerManager) UpdateTrackingList(hd []HostData) {
	log.Debug("--- Setting tracking list up")
	newHostList := make(map[string]HostData)
	for _, host := range hd {
		newHostList[host.CheckURL] = host
	}

	if hc.checker != nil {
		log.Debug("Reset initiated")
		hc.checker.ResetList(&newHostList)
	}
}

func InitHostCheckManager(store StorageHandler) {
	GlobalHostChecker = HostCheckerManager{}
	GlobalHostChecker.Init(store)
	GlobalHostChecker.Start()
}

func SetCheckerHostList() {
	log.Info("Loading uptime tests:")
	hostList := []HostData{}
	for _, spec := range ApiSpecRegister {
		for _, checkItem := range spec.UptimeTests.CheckList {
			hostList = append(hostList, GlobalHostChecker.PrepareTrackingHost(checkItem, spec.APIID))
			log.Info("---> Adding uptime test: ", checkItem.CheckURL)
		}
	}

	GlobalHostChecker.UpdateTrackingList(hostList)

	// Test fubctions
	// go func() {
	// 	time.Sleep(30 * time.Second)
	// 	isDown := GlobalHostChecker.IsHostDown("http://sharrow.tyk.io:3000/banana/phone")
	// 	log.Warning("IS IT DOWN? ", isDown)
	// 	time.Sleep(30 * time.Second)

	// 	isDown2 := GlobalHostChecker.IsHostDown("http://sharrow.tyk.io:3000/banana/phone")
	// 	log.Warning("IS IT DOWN Now? ", isDown2)

	// }()

}
