package main.com.iotracks.core_networking.main;

import com.iotracks.api.IOFabricClient;
import com.iotracks.api.listener.IOFabricAPIListener;
import io.netty.util.internal.StringUtil;
import main.com.iotracks.core_networking.comsat_client.ComSatClient;
import main.com.iotracks.core_networking.comsat_client.ComSatClientThreadFactory;
import main.com.iotracks.core_networking.utils.Certificate;
import main.com.iotracks.core_networking.utils.ContainerConfig;

import java.util.concurrent.ThreadFactory;
import java.util.logging.Logger;

/**
 * Created by saeid on 4/8/16.
 */
public class CoreNetworking {

    public static ContainerConfig config = null;
    public static String containerId = "";
    public static IOFabricClient ioFabricClient;
    private Logger log = Logger.getLogger(CoreNetworking.class.getName());
    private ComSatClient[] connections;
    private Certificate cert;
    private boolean connecting;
    private IOFabricAPIListener listener;

    public static void main(String[] args) throws Exception {
        CoreNetworking instance = new CoreNetworking();

        if (args.length > 0 && args[0].startsWith("--id=")) {
            CoreNetworking.containerId = args[0].substring(args[0].indexOf('=') + 1);
        } else {
            CoreNetworking.containerId = System.getenv("SELFNAME");
        }

        if (StringUtil.isNullOrEmpty(CoreNetworking.containerId)) {
            instance.log.warning("container id has not been set");
            instance.log.warning("use --id=XXXX parameter or set the id as SELFNAME=XXXX environment property");
            System.exit(1);
        }

        instance.init();
    }

    private void getCertificate() {
        String path = System.getProperty("user.dir") + System.getProperty("file.separator");
        cert = new Certificate(path + "intermediate.crt");
        if (cert == null) {
            log.warning("error importing certificate.");
            System.exit(1);
        }
    }

    private void makeConnections() {
        ThreadFactory threadFactory = new ComSatClientThreadFactory();
        int connectionCount = CoreNetworking.config.getConnectionCount();
        connections = new ComSatClient[connectionCount];
        for (int counter = 0; counter < connectionCount; counter++) {
            while (isConnecting()) {
                try {
                    Thread.sleep(100);
                } catch (Exception e) {
                }
            }
            connecting();
            connections[counter] = new ComSatClient(cert, this, counter);
            threadFactory.newThread(connections[counter]).start();
        }
    }

    private void closeAllConnections() {
        for (int counter = 0; counter < CoreNetworking.config.getConnectionCount(); counter++) {
            if (connections[counter] != null)
                connections[counter].close();
        }
    }

    private void start() {
        if (config.getMode().equals("private")) {
            ioFabricClient.openMessageWebSocket(listener);
        } else if (config.getMode().equals("public")) {
            config.setLocalHost("127.0.0.1");
            config.setLocalPort(8007);
        } else {
            while (config.getMode().equals("")) {
                try {
                    Thread.sleep(2000);
                } catch (Exception e){}
            }
        }

        makeConnections();

        while (true) {
            try {
                for (int counter = 0; counter < CoreNetworking.config.getConnectionCount(); counter++)
                    connections[counter].beat();
                Thread.sleep(CoreNetworking.config.getHeartbeatFrequency());
            } catch (Exception e) {
            }
        }
    }

    private void init() {
        getCertificate();

        String ioFabricHost = System.getProperty("iofabric_host", "localhost");
        int ioFabricPort = 54321;
        try {
            ioFabricPort = Integer.parseInt(System.getProperty("iofabric_port", "54321"));
        } catch (Exception e) {}
        ioFabricClient = new IOFabricClient(ioFabricHost, ioFabricPort, CoreNetworking.containerId);
        listener = new APIListenerImpl(this);
        ioFabricClient.fetchContainerConfig(listener);
        ioFabricClient.openControlWebSocket(listener);

        start();
    }

    public void connectingDone() {
        synchronized (CoreNetworking.class) {
            connecting = false;
        }
    }

    public boolean isConnecting() {
        synchronized (CoreNetworking.class) {
            return connecting;
        }
    }

    public void connecting() {
        synchronized (CoreNetworking.class) {
            connecting = true;
        }
    }

    public void setConfig(ContainerConfig config) {
        CoreNetworking.config = config;
    }

    public void updateConfig() {
        ioFabricClient.fetchContainerConfig(listener);
        closeAllConnections();
        init();
    }

}
