package main.com.iotracks.core_networking.main;

import com.iotracks.api.listener.IOFabricAPIListener;
import com.iotracks.elements.IOMessage;
import main.com.iotracks.core_networking.utils.ContainerConfig;
import main.com.iotracks.core_networking.utils.MessageRepository;

import javax.json.JsonObject;
import java.util.List;

/**
 * LocalApi listener
 *
 * Created by saeid on 4/8/16.
 */
public class APIListenerImpl implements IOFabricAPIListener {

    private final CoreNetworking coreNetworking;

    public APIListenerImpl(CoreNetworking coreNetworking) {
        this.coreNetworking = coreNetworking;
    }

    @Override
    public void onMessages(List<IOMessage> messages) {
        messages.forEach(message -> MessageRepository.pushMessage(message));
    }

    @Override
    public void onMessagesQuery(long timeframestart, long timeframeend, List<IOMessage> messages) {
        //
    }

    @Override
    public void onError(Throwable cause) {
        System.out.println("error");
    }

    @Override
    public void onBadRequest(String error) {
        System.out.println("bad request");
    }

    @Override
    public void onMessageReceipt(String messageId, long timestamp) {
    }

    @Override
    public void onNewConfig(JsonObject config) {
        coreNetworking.setConfig(new ContainerConfig(config));
    }

    @Override
    public void onNewConfigSignal() {
        coreNetworking.updateConfig();
    }
}
