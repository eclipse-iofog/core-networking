package main.com.iotracks.core_networking.utils;

import javax.json.JsonObject;

/**
 * Created by saeid on 4/8/16.
 */
public class ContainerConfig {
    private String mode;
    private String host;
    private int port;
    private int connectionCount;
    private String passCode;
    private String localHost;
    private int localPort;
    private int heartbeatFrequency;
    private int heartbeatThreshold;

    public ContainerConfig(JsonObject json) {
        setMode(json.containsKey("mode") ? json.getString("mode") : "");
        setHost(json.containsKey("host") ? json.getString("host") : "");
        setPort(json.containsKey("port") ? json.getInt("port") : 0);
        setConnectionCount(json.containsKey("connectioncount") ? json.getInt("connectioncount") : 0);
        setPassCode(json.containsKey("passcode") ? json.getString("passcode") : "");
        setLocalHost(json.containsKey("localhost") ? json.getString("localhost") : "");
        setLocalPort(json.containsKey("localport") ? json.getInt("localport") : 0);
        setHeartbeatFrequency(json.containsKey("heartbeatfrequency") ? json.getInt("heartbeatfrequency") : 0);
        setHeartbeatThreshold(json.containsKey("heartbeatabsencethreshold") ? json.getInt("heartbeatabsencethreshold") : 0);
    }

    public String getMode() {
        return mode;
    }

    public void setMode(String mode) {
        this.mode = mode;
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public int getConnectionCount() {
        return connectionCount;
    }

    public void setConnectionCount(int connectionCount) {
        this.connectionCount = connectionCount;
    }

    public String getPassCode() {
        return passCode;
    }

    public void setPassCode(String passCode) {
        this.passCode = passCode;
    }

    public String getLocalHost() {
        return localHost;
    }

    public void setLocalHost(String localHost) {
        this.localHost = localHost;
    }

    public int getLocalPort() {
        return localPort;
    }

    public void setLocalPort(int localPort) {
        this.localPort = localPort;
    }

    public int getHeartbeatFrequency() {
        return heartbeatFrequency;
    }

    public void setHeartbeatFrequency(int heartbeatFrequency) {
        this.heartbeatFrequency = heartbeatFrequency;
    }

    public int getHeartbeatThreshold() {
        return heartbeatThreshold;
    }

    public void setHeartbeatThreshold(int heartbeatThreshold) {
        this.heartbeatThreshold = heartbeatThreshold;
    }

    @Override
    public boolean equals(Object o) {
        ContainerConfig other = ((ContainerConfig) o);
        return other.getConnectionCount() == this.getConnectionCount() &&
                other.getHeartbeatFrequency() == this.getHeartbeatFrequency() &&
                other.getHeartbeatThreshold() == this.getHeartbeatThreshold() &&
                other.getHost().equals(this.getHost()) &&
                other.getPort() == this.getPort() &&
                other.getLocalHost().equals(this.getLocalHost()) &&
                other.getLocalPort() == this.getLocalPort() &&
                other.getMode().equals(this.getMode()) &&
                other.getPassCode().equals(this.getPassCode());
    }
}
