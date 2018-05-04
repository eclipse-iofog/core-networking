package main.org.eclipse.iofog.core_networking.utils;

import javax.json.JsonObject;

/**
 * class to hold container configuration
 * <p>
 * Created by saeid on 4/8/16.
 */
public class ContainerConfig {
    private final String MODE_FIELD_NAME = "mode";
    private final String HOST_FIELD_NAME = "host";
    private final String PORT_FIELD_NAME = "port";
    private final String CONNECTION_COUNT_FIELD_NAME = "connectioncount";
    private final String PASS_CODE_FIELD_NAME = "passcode";
    private final String LOCAL_HOST_FIELD_NAME = "localhost";
    private final String LOCAL_PORT_FIELD_NAME = "localport";
    private final String HEARTBEAT_FREQUENCY_FIELD_NAME = "heartbeatfrequency";
    private final String HEARTBEAT_THRESHOLD_FIELD_NAME = "heartbeatabsencethreshold";


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
        setMode(json.containsKey(MODE_FIELD_NAME) ? json.getString(MODE_FIELD_NAME) : "");
        setHost(json.containsKey(HOST_FIELD_NAME) ? json.getString(HOST_FIELD_NAME) : "");
        setPort(json.containsKey(PORT_FIELD_NAME) ? json.getInt(PORT_FIELD_NAME) : 0);
        setConnectionCount(json.containsKey(CONNECTION_COUNT_FIELD_NAME) ? json.getInt(CONNECTION_COUNT_FIELD_NAME) : 0);
        setPassCode(json.containsKey(PASS_CODE_FIELD_NAME) ? json.getString(PASS_CODE_FIELD_NAME) : "");
        setLocalHost(json.containsKey(LOCAL_HOST_FIELD_NAME) ? json.getString(LOCAL_HOST_FIELD_NAME) : "");
        setLocalPort(json.containsKey(LOCAL_PORT_FIELD_NAME) ? json.getInt(LOCAL_PORT_FIELD_NAME) : 0);
        setHeartbeatFrequency(json.containsKey(HEARTBEAT_FREQUENCY_FIELD_NAME) ? json.getInt(HEARTBEAT_FREQUENCY_FIELD_NAME) : 0);
        setHeartbeatThreshold(json.containsKey(HEARTBEAT_THRESHOLD_FIELD_NAME) ? json.getInt(HEARTBEAT_THRESHOLD_FIELD_NAME) : 0);
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
