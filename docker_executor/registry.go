package docker_executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type RegistryClient struct {
	Endpoint string
}

func (rc RegistryClient) getProcessorVersion(username string, name string, version string) (RegistryProcessorVersionRes, error) {
	url := rc.Endpoint + "/api/v1/Processor/slug/" + username + "/" + name + "/versions/" + version

	fmt.Println("🔍 Getting version of processor:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("🚨 Error occurred making a request to %s: %v\n", url, err)
		return RegistryProcessorVersionRes{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("🚨 Error reading response body: %v\n", err)
		return RegistryProcessorVersionRes{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🚨 Unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
		return RegistryProcessorVersionRes{}, fmt.Errorf("unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
	}

	var res RegistryProcessorVersionRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Printf("🚨 Error unmarshaling response to struct: %v\n", err)
		return RegistryProcessorVersionRes{}, err
	}
	return res, nil
}

func (rc RegistryClient) getProcessorVersionLatest(username string, name string) (RegistryProcessorVersionRes, error) {
	url := rc.Endpoint + "/api/v1/Processor/slug/" + username + "/" + name + "/versions/latest"

	fmt.Println("🔍 Getting latest version of processor:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("🚨 Error occurred making a request to %s: %v\n", url, err)
		return RegistryProcessorVersionRes{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("🚨 Error reading response body: %v\n", err)
		return RegistryProcessorVersionRes{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🚨 Unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
		return RegistryProcessorVersionRes{}, fmt.Errorf("unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
	}

	var res RegistryProcessorVersionRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Printf("🚨 Error unmarshaling response to struct: %v\n", err)
		return RegistryProcessorVersionRes{}, err
	}
	return res, nil
}

func (rc RegistryClient) getPluginVersion(username string, name string, version string) (RegistryPluginVersionRes, error) {
	url := rc.Endpoint + "/api/v1/Plugin/slug/" + username + "/" + name + "/versions/" + version

	fmt.Println("🔍 Getting version of plugin:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("🚨 Error occurred making a request to %s: %v\n", url, err)
		return RegistryPluginVersionRes{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("🚨 Error reading response body: %v\n", err)
		return RegistryPluginVersionRes{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🚨 Unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
		return RegistryPluginVersionRes{}, fmt.Errorf("unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
	}

	var res RegistryPluginVersionRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Printf("🚨 Error unmarshaling response to struct: %v\n", err)
		return RegistryPluginVersionRes{}, err
	}
	return res, nil
}

func (rc RegistryClient) getPluginVersionLatest(username string, name string) (RegistryPluginVersionRes, error) {
	url := rc.Endpoint + "/api/v1/Plugin/slug/" + username + "/" + name + "/versions/latest"

	fmt.Println("🔍 Getting latest version of plugin:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("🚨 Error occurred making a request to %s: %v\n", url, err)
		return RegistryPluginVersionRes{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("🚨 Error reading response body: %v\n", err)
		return RegistryPluginVersionRes{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🚨 Unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
		return RegistryPluginVersionRes{}, fmt.Errorf("unexpected status code: %d. Body: %s\n", resp.StatusCode, body)
	}

	var res RegistryPluginVersionRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Printf("🚨 Error unmarshaling response to struct: %v\n", err)
		return RegistryPluginVersionRes{}, err
	}
	return res, nil
}

func (rc RegistryClient) convertProcessor(cp CyanProcessorReq, processors []ProcessorRes) (CyanProcessor, error) {

	n := cp.Name
	username, name, version, err := parseCyanReference(n)
	if err != nil {
		return CyanProcessor{}, err
	}
	if version == nil {
		res, e := rc.getProcessorVersionLatest(username, name)
		if e != nil {
			fmt.Printf("🚨 Error getting latest version of processor %s/%s: %v\n", username, name, e)
			return CyanProcessor{}, e
		}
		for i := res.Principal.Version; i > 0; i-- {
			v := strconv.Itoa(i)
			r, er := rc.getProcessorVersion(username, name, v)
			if er != nil {
				fmt.Printf("🚨 Error getting latest version of processor %s/%s:%s %v\n", username, name, v, e)
				return CyanProcessor{}, e
			}
			for _, p := range processors {
				if p.ID == r.Principal.Id {
					fmt.Printf("✅ Processor %s (from Cyan Response)'s verions %s matches %s", n, v, p.ID)
					return CyanProcessor{
						Id:        r.Principal.Id,
						Reference: n,
						Username:  username,
						Name:      name,
						Version:   res.Principal.Version,
						Config:    cp.Config,
						Files:     cp.Files,
					}, nil
				}
			}
			fmt.Printf("⚠️ Processor %s (from Cyan Response)'s verions %s does not match any processor defined in Template", n, v)
		}
		er := fmt.Errorf("processor %s (from Cyan Response) does not have a matching version defined in the template", n)
		fmt.Printf("🚨 Processor %s (from Cyan Response) does not have a matching version defined in the template", n)
		return CyanProcessor{}, er
	} else {
		v := *version
		res, e := rc.getProcessorVersion(username, name, v)
		if e != nil {
			fmt.Printf("🚨 Error getting version of processor %s/%s:%s: %v\n", username, name, v, e)
			return CyanProcessor{}, e
		}
		return CyanProcessor{
			Id:        res.Principal.Id,
			Reference: n,
			Username:  username,
			Name:      name,
			Version:   res.Principal.Version,
			Config:    cp.Config,
			Files:     cp.Files,
		}, nil
	}
}

func (rc RegistryClient) convertPlugin(cp CyanPluginReq, plugins []PluginRes) (CyanPlugin, error) {
	n := cp.Name
	username, name, version, err := parseCyanReference(n)
	if err != nil {
		return CyanPlugin{}, err
	}
	if version == nil {
		res, e := rc.getPluginVersionLatest(username, name)
		if e != nil {
			fmt.Printf("🚨 Error getting latest version of plugin %s/%s: %v\n", username, name, e)
			return CyanPlugin{}, e
		}
		for i := res.Principal.Version; i > 0; i-- {
			v := strconv.Itoa(i)
			r, er := rc.getPluginVersion(username, name, v)
			if er != nil {
				fmt.Printf("🚨 Error getting latest version of plugin %s/%s:%s %v\n", username, name, v, e)
				return CyanPlugin{}, e
			}
			for _, p := range plugins {
				if p.ID == r.Principal.Id {
					fmt.Printf("✅ Plugin %s (from Cyan Response)'s verions %s matches %s", n, v, p.ID)
					return CyanPlugin{
						Id:        res.Principal.Id,
						Reference: n,
						Username:  username,
						Name:      name,
						Version:   res.Principal.Version,
						Config:    cp.Config,
					}, nil
				}
			}
			fmt.Printf("⚠️ Plugin %s (from Cyan Response)'s verions %s does not match any plugin defined in Template", n, v)
		}
		er := fmt.Errorf("plugin %s (from Cyan Response) does not have a matching version defined in the template", n)
		fmt.Printf("🚨 plugin %s (from Cyan Response) does not have a matching version defined in the template", n)
		return CyanPlugin{}, er
	} else {
		v := *version
		res, e := rc.getPluginVersion(username, name, v)
		if e != nil {
			fmt.Printf("🚨 Error getting version of plugin %s/%s:%s: %v\n", username, name, v, e)
			return CyanPlugin{}, e
		}
		return CyanPlugin{
			Id:        res.Principal.Id,
			Reference: n,
			Username:  username,
			Name:      name,
			Version:   res.Principal.Version,
			Config:    cp.Config,
		}, nil
	}
}
