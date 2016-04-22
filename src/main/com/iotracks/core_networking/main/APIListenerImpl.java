package main.com.iotracks.core_networking.main;

import com.iotracks.api.listener.IOFabricAPIListener;
import com.iotracks.elements.IOMessage;
import main.com.iotracks.core_networking.utils.ContainerConfig;
import main.com.iotracks.core_networking.utils.MessageRepository;

import javax.json.JsonObject;
import java.util.List;
import java.util.logging.Logger;

/**
 * LocalApi listener
 * <p>
 * Created by saeid on 4/8/16.
 */
public class APIListenerImpl implements IOFabricAPIListener {

    private final Logger log = Logger.getLogger(APIListenerImpl.class.getName());
    private final CoreNetworking coreNetworking;

    public APIListenerImpl(CoreNetworking coreNetworking) {
        this.coreNetworking = coreNetworking;
    }

    @Override
    public void onMessages(List<IOMessage> messages) {
        log.info(String.format("%d new message(s) received", messages.size()));
        messages.forEach(message -> MessageRepository.pushMessage(message));
    }

    @Override
    public void onMessagesQuery(long timeframestart, long timeframeend, List<IOMessage> messages) {
    }

    @Override
    public void onError(Throwable cause) {
        log.warning(cause.getMessage());
    }

    @Override
    public void onBadRequest(String error) {
        log.warning(error);
    }

    @Override
    public void onMessageReceipt(String messageId, long timestamp) {
    }

    @Override
    public void onNewConfig(JsonObject config) {
        log.info("config received: \n" + config.toString());
        coreNetworking.setConfig(new ContainerConfig(config));
    }

    @Override
    public void onNewConfigSignal() {
        coreNetworking.updateConfig();
    }
}
